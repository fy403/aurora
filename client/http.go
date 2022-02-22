package client

import "time"

type AuthRequest struct {
	Name     string `json:"name" validate:"required,gt=0,lt=15"`
	Password string `json:"password" validate:"required,gt=0,lt=30"`
}

type AuthResponse struct {
	Message string `json:"message"`
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
}

type CenterRequest struct {
	UUID            string       `json:"uuid"`    // user id
	User            string       `json:"user"`    // user name
	BatchID         string       `json:"batchId"` // unique id for request
	Timestamp       int64        `json:"timestamp" validate:"required"`
	TaskType        string       `json:"taskType" validate:"required,oneof='task' 'group' 'chord' 'chain'"`
	Signatures      []*Signature `json:"signatures" validate:"required,gt=0"`
	SleepDuration   int          `json:"sleepDuration" validate:"required,min=5,max=5000"` // min=5ms max=5s
	SendConcurrency int          `json:"sendConcurrency" validate:"min=0,max=10"`
	CallBack        *Signature   `json:"callBack" validate:"required_if=TaskType chord"`
}

type CenterResponse struct {
	UUID          string          `json:"uuid"` // user id
	User          string          `json:"user"`
	BatchID       string          `json:"batchId"` // unique id for request
	Timestamp     int64           `json:"timestamp"`
	TaskType      string          `json:"taskType"`
	TaskResponses []*TaskResponse `json:"taskResponse"`
}

type TaskResponse struct {
	Results    []interface{} `json:"results"`
	Signatures []*Signature  `json:"signatures"`
	CallBack   *Signature    `json:"callBack"`
}

// Signature represents a single task invocation
type Signature struct {
	UUID           string
	Name           string
	RoutingKey     string
	ETA            *time.Time
	GroupUUID      string
	GroupTaskCount int
	Args           []Arg
	Headers        Headers
	Priority       uint8
	Immutable      bool
	RetryCount     int
	RetryTimeout   int
	OnSuccess      []*Signature
	OnError        []*Signature
	ChordCallback  *Signature
	//MessageGroupId for Broker, e.g. SQS
	BrokerMessageGroupId string
	//ReceiptHandle of SQS Message
	SQSReceiptHandle string
	// StopTaskDeletionOnError used with sqs when we want to send failed messages to dlq,
	// and don't want aurora to delete from source queue
	StopTaskDeletionOnError bool
	// IgnoreWhenTaskNotRegistered auto removes the request when there is no handeler available
	// When this is true a task with no handler will be ignored and not placed back in the queue
	IgnoreWhenTaskNotRegistered bool
}

type Arg struct {
	Name  string      `bson:"name"`
	Type  string      `bson:"type"`
	Value interface{} `bson:"value"`
}

// Headers represents the headers which should be used to direct the task
type Headers map[string]interface{}
