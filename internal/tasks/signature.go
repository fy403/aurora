package tasks

import (
	"aurora/internal/utils"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Arg represents a single argument passed to invocation fo a task
type Arg struct {
	// Name  string      `json:"name" bson:"name"`
	Type  string      `json:"type" bson:"type"`
	Value interface{} `json:"value" bson:"value"`
}

// Headers represents the headers which should be used to direct the task
type Headers map[string]interface{}

// Set on Headers implements opentracing.TextMapWriter for trace propagation
func (h Headers) Set(key, val string) {
	h[key] = val
}

// ForeachKey on Headers implements opentracing.TextMapReader for trace propagation.
// It is essentially the same as the opentracing.TextMapReader implementation except
// for the added casting from interface{} to string.
func (h Headers) ForeachKey(handler func(key, val string) error) error {
	for k, v := range h {
		// Skip any non string values
		stringValue, ok := v.(string)
		if !ok {
			continue
		}

		if err := handler(k, stringValue); err != nil {
			return err
		}
	}

	return nil
}

// Signature represents a single task invocation
type Signature struct {
	ID             string            `json:"id"`
	UUID           string            `json:"uuid"`
	Name           string            `json:"name"`
	RoutingKey     string            `json:"routing_key"`
	LabelSelector  map[string]string `json:"label_selector"`
	ETA            *time.Time        `json:"eta"`
	GroupUUID      string            `json:"group_uuid"`
	GroupTaskCount int               `json:"group_task_count"`
	GraphUUID      string            `json:"graph_uuid"`
	GraphTaskCount int               `json:"graph_task_count"`
	Args           []Arg             `json:"args"`
	Headers        Headers           `json:"headers"`
	Priority       uint8             `json:"priority"`
	Immutable      bool              `json:"immutable"`
	RetryCount     int               `json:"retry_count"`
	RetryTimeout   int               `json:"retry_timeout"`
	OnSuccess      []*Signature      `json:"on_success"`
	OnError        []*Signature      `json:"on_error"`
	ChordCallback  *Signature        `json:"chord_callback"`
	//MessageGroupId for Broker, e.g. SQS
	BrokerMessageGroupId string `json:"broker_message_group_id"`
	//ReceiptHandle of SQS Message
	SQSReceiptHandle string `json:"sqs_receipt_handle"`
	// StopTaskDeletionOnError used with sqs when we want to send failed messages to dlq,
	// and don't want aurora to delete from source queue
	StopTaskDeletionOnError bool `json:"stop_task_deletion_on_error"`
	// IgnoreWhenTaskNotRegistered auto removes the request when there is no handeler available
	// When this is true a task with no handler will be ignored and not placed back in the queue
	IgnoreWhenTaskNotRegistered bool `json:"ignore_when_task_not_registered"`
}

// NewSignature creates a new task signature
func NewSignature(name string, args []Arg) (*Signature, error) {
	signatureID := uuid.New().String()
	return &Signature{
		UUID: fmt.Sprintf("task_%v", signatureID),
		Name: name,
		Args: args,
	}, nil
}

func CopySignatures(signatures ...*Signature) []*Signature {
	var sigs = make([]*Signature, len(signatures))
	for index, signature := range signatures {
		sigs[index] = CopySignature(signature)
	}
	return sigs
}

func CopySignature(signature *Signature) *Signature {
	var sig = new(Signature)
	_ = utils.DeepCopy(sig, signature)
	return sig
}
