package api

import (
	"aurora/internal/backends/result"
	"aurora/internal/log"
	"aurora/internal/request"
	"aurora/internal/tasks"
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	opentracing_log "github.com/opentracing/opentracing-go/log"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
)

type taskHandler struct{}

func (*taskHandler) send(wait *WaitConn, req request.CenterRequest) {
	log.Runtime().Infof("%s %v", wait.GetRoute(), req)
	defer func() { wait.Done() }()

	if err := req.Validate(); err != nil {
		wait.SetResult(fmt.Sprintf("HTTPTasksSend Failed to validate format: %#v", req), "")
		return
	}

	if err := req.Inject(wait.ctx.Request); err != nil {
		wait.SetResult(err.Error(), "")
		return
	}

	if err := defaultApi.LabelSelector(&req); err != nil {
		wait.SetResult(err.Error(), "")
		return
	}

	/*
	 * Lets start a span representing defaultApi run of the `send` command and
	 * set a batch id as baggage so it can travel all the way into
	 * the worker functions.
	 */
	span, ctx := opentracing.StartSpanFromContext(context.Background(), "send")
	defer span.Finish()

	// Generate an unique id for each request
	batchID := uuid.New().String()
	// sync server config
	req.BatchID = batchID
	// Local Span details
	span.SetTag("task.type", req.TaskType)
	span.LogFields(opentracing_log.String("params.format", fmt.Sprintf("%#v", req)))
	span.LogFields(opentracing_log.String("batch.id", batchID))
	// Span Context：pass across process boundary
	span.SetBaggageItem("user.uuid", req.UUID)
	span.SetBaggageItem("user.name", req.User)
	span.SetBaggageItem("batch.id", batchID)

	log.Runtime().Infof("Starting batch: %s", batchID)
	time.Local, _ = time.LoadLocation("Asia/Beijing")

	responseOBJ := request.CenterResponse{
		UUID:          req.UUID,
		User:          req.User,
		BatchID:       req.BatchID,
		Timestamp:     time.Now().Local().Unix(),
		TaskType:      req.TaskType,
		TaskResponses: []*request.TaskResponse{},
	}

	log.Runtime().Debugf("Received A Send : %s, %s", req.TaskType, req.BatchID)

	switch v := req.TaskType; v {
	case "task":
		asyncResultPtr, err := defaultApi.server.SendTaskWithContext(ctx, req.Signatures[0])
		if err != nil {
			wait.SetResult(fmt.Sprintf("Could not send task: %s", err.Error()), "")
			return
		}
		// Try to obtain results,In time limit
		results, err := asyncResultPtr.GetWithTimeout(time.Duration(req.TimeoutDuration)*time.Millisecond, time.Duration(req.SleepDuration)*time.Millisecond)
		if err != nil && err != result.ErrTimeoutReached {
			wait.SetResult(fmt.Sprintf("Task has failed with error: %s", err.Error()), "")
			return
		}
		// Clean sensitive information
		tasks.CleanSignatureSensitiveInfo(asyncResultPtr.Signature)
		responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
			Results: tasks.InterfaceReadableResults(results),
			Signatures: []*tasks.Signature{
				asyncResultPtr.Signature,
			},
		})
	case "group":
		group, err := tasks.NewGroup(req.Signatures...)
		if err != nil {
			wait.SetResult(fmt.Sprintf("Error creating group: %s", err.Error()), "")
			return
		}

		asyncResults, err := defaultApi.server.SendGroupWithContext(ctx, group, req.SendConcurrency)
		if err != nil {
			wait.SetResult(fmt.Sprintf("Could not send group: %s", err.Error()), "")
			return
		}
		size := len(asyncResults)
		for _, asyncResultPtr := range asyncResults {
			// Try to obtain results,In time limit
			results, err := asyncResultPtr.GetWithTimeout(time.Duration(req.TimeoutDuration/size)*time.Millisecond, time.Duration(req.SleepDuration)*time.Millisecond)
			if err != nil && err != result.ErrTimeoutReached {
				wait.SetResult(fmt.Sprintf("Task has failed with error: %s", err.Error()), "")
				return
			}
			// Clean sensitive information
			tasks.CleanSignatureSensitiveInfo(asyncResultPtr.Signature)
			responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
				Results: tasks.InterfaceReadableResults(results),
				Signatures: []*tasks.Signature{
					asyncResultPtr.Signature,
				},
			})
		}
	case "graph":
		graph, err := tasks.NewGraph(req.Relations, req.Signatures...)
		if err != nil {
			wait.SetResult(fmt.Sprintf("Error creating graph: %s", err.Error()), "")
			return
		}
		asyncResults, err := defaultApi.server.SendGraphWithContext(ctx, graph)
		if err != nil {
			wait.SetResult(fmt.Sprintf("Could not send graph: %s", err.Error()), "")
			return
		}
		size := len(asyncResults)
		for _, asyncResultPtr := range asyncResults {
			// Try to obtain results,In time limit
			results, err := asyncResultPtr.GetWithTimeout(time.Duration(req.TimeoutDuration/size)*time.Millisecond, time.Duration(req.SleepDuration)*time.Millisecond)
			if err != nil && err != result.ErrTimeoutReached {
				wait.SetResult(fmt.Sprintf("Task has failed with error: %s", err.Error()), "")
			}
			// Clean sensitive information
			tasks.CleanSignatureSensitiveInfo(asyncResultPtr.Signature)
			responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
				Results: tasks.InterfaceReadableResults(results),
				Signatures: []*tasks.Signature{
					asyncResultPtr.Signature,
				},
			})
		}
	case "chord":
		group, err := tasks.NewGroup(req.Signatures...)
		if err != nil {
			wait.SetResult(fmt.Sprintf("Error creating group: %s", err.Error()), "")
			return
		}

		chord, err := tasks.NewChord(group, req.CallBack)
		if err != nil {
			wait.SetResult(fmt.Sprintf("Error creating chord: %s", err.Error()), "")
			return
		}

		chordAsyncResult, err := defaultApi.server.SendChordWithContext(ctx, chord, req.SendConcurrency)
		if err != nil {
			wait.SetResult(fmt.Sprintf("Could not send chord: %s", err.Error()), "")
			return
		}

		var signatures []*tasks.Signature
		for _, asyncResultPtr := range chordAsyncResult.GetGroupAsyncResults() {
			// Clean sensitive information
			tasks.CleanSignatureSensitiveInfo(asyncResultPtr.Signature)
			signatures = append(signatures, asyncResultPtr.Signature)
		}
		// Try to obtain results,In time limit
		results, err := chordAsyncResult.GetWithTimeout(time.Duration(req.TimeoutDuration)*time.Millisecond, time.Duration(req.SleepDuration)*time.Millisecond)
		if err != nil && err != result.ErrTimeoutReached {
			wait.SetResult(fmt.Sprintf("Task has failed with error: %s", err.Error()), "")
			return
		}
		responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
			Results:    tasks.InterfaceReadableResults(results),
			Signatures: signatures,
			CallBack:   chordAsyncResult.GetChordAyncResults().Signature,
		})
	case "chain":
		chain, err := tasks.NewChain(req.Signatures...)
		if err != nil {
			wait.SetResult(fmt.Sprintf("Error creating chain: %s", err), "")
			return
		}

		chainAsyncResult, err := defaultApi.server.SendChainWithContext(ctx, chain)
		if err != nil {
			wait.SetResult(fmt.Sprintf("Could not send chain: %s", err.Error()), "")
			return
		}

		var signatures []*tasks.Signature
		for _, asyncResultPtr := range chainAsyncResult.GetAsyncResults() {
			// Clean sensitive information
			tasks.CleanSignatureSensitiveInfo(asyncResultPtr.Signature)
			signatures = append(signatures, asyncResultPtr.Signature)
		}
		// Try to obtain results,In time limit
		results, err := chainAsyncResult.GetWithTimeout(time.Duration(req.TimeoutDuration)*time.Millisecond, time.Duration(req.SleepDuration)*time.Millisecond)
		if err != nil && err != result.ErrTimeoutReached {
			wait.SetResult(fmt.Sprintf("Task has failed with error: %s", err.Error()), "")
		}
		responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
			Results:    tasks.InterfaceReadableResults(results),
			Signatures: signatures,
		})
	default:
		err := errors.New("Unexpected task type: " + v)
		span.SetTag("error", true)
		span.LogFields(
			opentracing_log.Error(err),
		)
		wait.SetResult(err.Error(), "")
		return
	}

	hasFinished := true
	// If success, the results is not nil but may be empty
	for _, tasksResponse := range responseOBJ.TaskResponses {
		if tasksResponse.Results == nil {
			hasFinished = false
			break
		}
	}
	if !hasFinished {
		wait.SetResult("continue", responseOBJ)
	} else {
		wait.SetResult("", responseOBJ)
	}
}

func (*taskHandler) touch(wait *WaitConn, req request.CenterRequest) {
	log.Runtime().Infof("%s %v", wait.GetRoute(), req)
	defer func() { wait.Done() }()

	var err error
	if err = req.Validate(); err != nil {
		wait.SetResult(fmt.Sprintf("HTTPTasksTouch Failed to validate format: %#v", req), "")
		return
	}

	if err = req.Inject(wait.ctx.Request); err != nil {
		wait.SetResult(err.Error(), "")
		return
	}

	/*
	* Send the corresponding query according to the TaskType matching result
	 */
	responseOBJ := request.CenterResponse{
		UUID:          req.UUID,
		User:          req.User,
		BatchID:       req.BatchID,
		Timestamp:     time.Now().Local().Unix(),
		TaskType:      req.TaskType,
		TaskResponses: []*request.TaskResponse{},
	}
	hasFinished := true

	log.Runtime().Debugf("Received A Touch: %s, %s", req.TaskType, req.BatchID)

	switch v := req.TaskType; v {
	case "task":
		asyncResult := result.NewAsyncResult(req.Signatures[0], defaultApi.server.GetBackend())
		if !asyncResult.GetState().IsSuccess() {
			hasFinished = false
			break
		}
		results, err := asyncResult.Touch()
		if err != nil {
			wait.SetResult(fmt.Sprintf("Task has failed with error: %s", err.Error()), "")
			return
		}
		responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
			Results:    tasks.InterfaceReadableResults(results),
			Signatures: req.Signatures,
		})
	case "group", "graph":
		for _, signature := range req.Signatures {
			asyncResult := result.NewAsyncResult(signature, defaultApi.server.GetBackend())
			if !asyncResult.GetState().IsSuccess() {
				hasFinished = false
				break
			}
			results, err := asyncResult.Touch()
			if err != nil {
				wait.SetResult(fmt.Sprintf("Task has failed with error: %s", err.Error()), "")
				return
			}
			responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
				Results: tasks.InterfaceReadableResults(results),
				Signatures: []*tasks.Signature{
					signature,
				},
			})
		}
	case "chord":
		chordAsyncResult := result.NewChordAsyncResult(req.Signatures, req.CallBack, defaultApi.server.GetBackend())
		for _, asyncResult := range chordAsyncResult.GetGroupAsyncResults() {
			if !asyncResult.GetState().IsSuccess() {
				hasFinished = false
				break
			}
			_, err := asyncResult.Touch()
			if err != nil {
				wait.SetResult(fmt.Sprintf("Task has failed with error: %s", err.Error()), "")
				return
			}
		}
		if !chordAsyncResult.GetChordAyncResults().GetState().IsSuccess() {
			hasFinished = false
			break
		}
		results, err := chordAsyncResult.GetChordAyncResults().Touch()
		if err != nil {
			wait.SetResult(fmt.Sprintf("Task has failed with error: %s", err.Error()), "")
			return
		}
		responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
			Results:    tasks.InterfaceReadableResults(results),
			Signatures: req.Signatures,
			CallBack:   req.CallBack,
		})
	case "chain":
		chainAsyncResult := result.NewChainAsyncResult(req.Signatures, defaultApi.server.GetBackend())
		var results []reflect.Value
		for _, asyncResult := range chainAsyncResult.GetAsyncResults() {
			if !asyncResult.GetState().IsSuccess() {
				hasFinished = false
				break
			}
			results, err = asyncResult.Touch()
			if err != nil {
				wait.SetResult(fmt.Sprintf("Task has failed with error: %s", err.Error()), "")
				return
			}
		}
		if !hasFinished {
			break
		}
		responseOBJ.TaskResponses = append(responseOBJ.TaskResponses, &request.TaskResponse{
			Results:    tasks.InterfaceReadableResults(results),
			Signatures: req.Signatures,
		})
	default:
		err := errors.New("Unexpected task type: " + v)
		wait.SetResult(err.Error(), "")
		return

	}
	if !hasFinished {
		wait.SetResult("continue", responseOBJ)
	} else {
		wait.SetResult("", responseOBJ)
	}
}
