package request

import (
	"aurora/internal/tasks"
)

type CenterRequest struct {
	UUID            string             `json:"uuid"`     // user id
	User            string             `json:"user"`     // user name
	BatchID         string             `json:"batch_id"` // unique id for request
	Timestamp       int64              `json:"timestamp" validate:"required"`
	TaskType        string             `json:"task_type" validate:"required,oneof='task' 'group' 'chord' 'chain' 'graph'"`
	LabelSelector   map[string]string  `json:"label_selector"`
	Relations       []map[int]int      `json:"relations"`
	Signatures      []*tasks.Signature `json:"signatures" validate:"required,gt=0"`
	TimeoutDuration int                `json:"timeout_duration" validate:"min=0"`
	SleepDuration   int                `json:"sleep_duration" validate:"min=0"`
	SendConcurrency int                `json:"send_concurrency" validate:"min=0"`
	CallBack        *tasks.Signature   `json:"callBack" validate:"required_if=TaskType chord"`
}

type CenterResponse struct {
	UUID          string          `json:"uuid"` // user id
	User          string          `json:"user"`
	BatchID       string          `json:"batch_id"` // unique id for request
	Timestamp     int64           `json:"timestamp"`
	TaskType      string          `json:"task_type"`
	TaskUUID      string          `json:"task_uuid"`
	TaskResponses []*TaskResponse `json:"task_responses"`
}

type TaskResponse struct {
	Results    []interface{}      `json:"results"`
	State      string             `json:"state"`
	Signatures []*tasks.Signature `json:"signatures"`
	CallBack   *tasks.Signature   `json:"callBack"`
}

type WorkerRequest struct {
	UUID      string            `json:"uuid"`
	SpecQueue string            `json:"spec_queue"`
	Metrics   map[string]string `json:"metrics"`
	Handlers  []*Handler        `json:"handlers"`
	Labels    map[string]string `json:"labels"`
	Timestamp int64             `json:"timestamp"`
}

type WorkerResponse struct {
	UUID      string            `json:"uuid"`
	SpecQueue string            `json:"spec_queue"`
	Metrics   map[string]string `json:"metrics"`
	Handlers  []*Handler        `json:"handlers"`
	Labels    map[string]string `json:"labels"`
	Timestamp int64             `json:"timestamp"`
}

type WorkerMeta struct {
	UUID      string            `bson:"_id" json:"_id"`
	SpecQueue string            `bson:"spec_queue" json:"spec_queue"`
	Metrics   map[string]string `bson:"metrics" json:"metrics"`
	Handlers  []*Handler        `bson:"handlers" json:"handlers"`
	Labels    map[string]string `bson:"labels" json:"labels"`
	CreatedAt int64             `bson:"created_at" json:"created_at"`
}

type Handler struct {
	Name    string      `json:"name" bson:"name"`
	Usage   string      `json:"usage" bson:"usage"`
	InArgs  []tasks.Arg `json:"in_args" bson:"in_args"`
	OutArgs []tasks.Arg `json:"out_args" bson:"out_args"`
	Fn      interface{} `json:"-" bson:"-"`
}

type RabbitMQApi struct {
	Consumers int    `json:"consumers"`
	State     string `json:"state"`
	Messages  int    `json:"messages"`
}

type FaasResponse struct {
	FunctionID   string `json:"function_id"`
	Driver       string `json:"driver"`
	FunctionName string `json:"function_name"`
	Description  string `json:"description"`
	Timestamp    int64  `json:"timestamp"`
}
