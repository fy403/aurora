package request

import (
	"aurora/internal/tasks"
)

type CenterRequest struct {
	UUID            string             `json:"UUID" validate:"max=50"`    // user id
	User            string             `json:"User" validate:"max=15"`    // user name
	BatchID         string             `json:"BatchID" validate:"max=50"` // unique id for request
	Timestamp       int64              `json:"Timestamp" validate:"required"`
	TaskType        string             `json:"TaskType" validate:"required,oneof='task' 'group' 'chord' 'chain'"`
	LabelSelector   map[string]string  `json:"LabelSelector" validate:"max=128"`
	Signatures      []*tasks.Signature `json:"Signatures" validate:"required,gt=0"`
	TimeoutDuration int                `json:"TimeoutDuration" validate:"required,min=100,max=5000"`
	SleepDuration   int                `json:"SleepDuration" validate:"required,min=50,max=500"`
	SendConcurrency int                `json:"SendConcurrency" validate:"min=0,max=10"`
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
