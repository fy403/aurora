package request

import (
	"aurora/internal/tasks"
)

type CenterRequest struct {
	UUID            string             `json:"UUID"`    // user id
	User            string             `json:"User"`    // user name
	BatchID         string             `json:"BatchID"` // unique id for request
	Timestamp       int64              `json:"Timestamp" validate:"required"`
	TaskType        string             `json:"TaskType" validate:"required,oneof='task' 'group' 'chord' 'chain' 'graph'"`
	LabelSelector   map[string]string  `json:"LabelSelector"`
	Relations       []map[int]int      `json:"Relations"`
	Signatures      []*tasks.Signature `json:"Signatures" validate:"required,gt=0"`
	TimeoutDuration int                `json:"TimeoutDuration" validate:"min=0`
	SleepDuration   int                `json:"SleepDuration" validate:"min=0"`
	SendConcurrency int                `json:"SendConcurrency" validate:"min=0"`
	CallBack        *tasks.Signature   `json:"CallBack" validate:"required_if=TaskType chord"`
}

type CenterResponse struct {
	UUID          string          `json:"UUID"` // user id
	User          string          `json:"User"`
	BatchID       string          `json:"BatchID"` // unique id for request
	Timestamp     int64           `json:"Timestamp"`
	TaskType      string          `json:"TaskType"`
	TaskResponses []*TaskResponse `json:"TaskResponses"`
}

type TaskResponse struct {
	Results    []interface{}      `json:"Results"`
	Signatures []*tasks.Signature `json:"Signatures"`
	CallBack   *tasks.Signature   `json:"CallBack"`
}

type WorkerRequest struct {
	UUID      string            `json:"UUID"`
	SpecQueue string            `json:"SpecQueue"`
	Metrics   map[string]string `json:"Metrics"`
	Labels    map[string]string `json:"Labels"`
	Timestamp int64             `json:"Timestamp"`
}

type WorkerResponse struct {
	UUID      string            `json:"UUID"`
	SpecQueue string            `json:"SpecQueue"`
	Metrics   map[string]string `json:"Metrics"`
	Labels    map[string]string `json:"Labels"`
	Timestamp int64             `json:"Timestamp"`
}

type RabbitMQApi struct {
	Consumers int    `json:"consumers"`
	State     string `json:"state"`
}
