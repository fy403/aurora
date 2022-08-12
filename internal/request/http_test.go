package request_test

import (
	"aurora/internal/request"
	"aurora/internal/tasks"
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	obj := &request.CenterRequest{
		UUID:      "2343543543",
		User:      "admin",
		BatchID:   "1234324",
		Timestamp: time.Now().Unix(),
		TaskType:  "chord",
		Signatures: []*tasks.Signature{
			&addTask0,
			&addTask1,
		},
		SleepDuration:   5,
		SendConcurrency: 1,
		CallBack:        &multiplyTask1,
	}
	err := obj.Validate()
	assert.NoError(t, err)
}

func TestDecoder(t *testing.T) {
	// req, err := http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	initTasks()
	obj := request.CenterRequest{
		UUID:      "2343543543",
		User:      "admin",
		BatchID:   "1234324",
		Timestamp: time.Now().Unix(),
		TaskType:  "chord",
		Signatures: []*tasks.Signature{
			&addTask0,
			&addTask1,
		},
		SleepDuration:   5,
		SendConcurrency: 10,
		CallBack:        &multiplyTask1,
	}

	// Union Global Encode
	v, err := json.Marshal(obj)
	t.Logf("%s", v)
	assert.NoError(t, err)

	var requestOBJ request.CenterRequest

	// Union Global Decode
	dec := json.NewDecoder(bytes.NewReader(v))
	dec.UseNumber()

	err = dec.Decode(&requestOBJ)
	assert.NoError(t, err)
	// For interface of intx/uintx/floatx will be decoded as json.Number instead of default float64
	// So as to avoid lost precision
	// So there will be some inconsistencies here
	// But not influent result! For more information, please see: aurora/internal/tasks/reflect.go
	assert.Equal(t, &obj, &requestOBJ)
}

func TestIsValid(t *testing.T) {
	if os.Getenv("BROKER_API") == "" {
		t.Skip("BROKER_API is not defined")
	}
	resp := &request.WorkerResponse{
		UUID: "spec_queue:4211021996",
	}
	isValid := resp.IsValid(os.Getenv("BROKER_API"))
	assert.True(t, isValid)
}
