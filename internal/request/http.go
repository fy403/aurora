package request

import (
	"aurora/internal/tasks"

	validator "github.com/go-playground/validator/v10"
)

type CenterRequest struct {
	UUID            string             `json:"uuid" validate:"required"` // user id
	BatchID         string             `json:"batchId"`                  // unique id for request
	Timestamp       int64              `json:"timestamp" validate:"required"`
	TaskType        string             `json:"taskType" validate:"required,oneof='task' 'group' 'chord' 'chain'"`
	Signatures      []*tasks.Signature `json:"signatures" validate:"required,gt=0"`
	SleepDuration   int                `json:"sleepDuration" validate:"required,min=5,max=5000"` // min=5ms max=5s
	SendConcurrency int                `json:"sendConcurrency" validate:"min=0,max=10"`
	CallBack        *tasks.Signature   `json:"callBack" validate:"required_if=TaskType chord"`
}

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

func init() {
	validate = validator.New()
}

func (f *CenterRequest) Validate() (err error) {
	return validate.Struct(f)
}

type CenterResponse struct {
	UUID          string          `json:"uuid"`    // user id
	BatchID       string          `json:"batchId"` // unique id for request
	Timestamp     int64           `json:"timestamp"`
	TaskType      string          `json:"taskType"`
	TaskResponses []*TaskResponse `json:"taskResponse"`
}

type TaskResponse struct {
	Results    []interface{}      `json:"results"`
	Signatures []*tasks.Signature `json:"signatures"`
	CallBack   *tasks.Signature   `json:"callBack"`
}
