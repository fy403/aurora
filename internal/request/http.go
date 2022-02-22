package request

import (
	"aurora/internal/auth"
	"aurora/internal/tasks"
	"errors"
	"net/http"

	validator "github.com/go-playground/validator/v10"
)

type CenterRequest struct {
	UUID            string             `json:"uuid"`    // user id
	User            string             `json:"user"`    // user name
	BatchID         string             `json:"batchId"` // unique id for request
	Timestamp       int64              `json:"timestamp" validate:"required"`
	TaskType        string             `json:"taskType" validate:"required,oneof='task' 'group' 'chord' 'chain'"`
	Signatures      []*tasks.Signature `json:"signatures" validate:"required,gt=0"`
	SleepDuration   int                `json:"sleepDuration" validate:"required,min=5,max=5000"` // min=5ms max=5s
	SendConcurrency int                `json:"sendConcurrency" validate:"min=0,max=10"`
	CallBack        *tasks.Signature   `json:"callBack" validate:"required_if=TaskType chord"`
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
	Results    []interface{}      `json:"results"`
	Signatures []*tasks.Signature `json:"signatures"`
	CallBack   *tasks.Signature   `json:"callBack"`
}

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Verification parameters
func (f *CenterRequest) Validate() (err error) {
	if err = validate.Struct(f); err != nil {
		return err
	}
	return nil
}

// Inject will add some session attribute in f
func (f *CenterRequest) Inject(r *http.Request) (err error) {
	// Inject session`s attribute in f
	store := auth.DefaultStore()
	if store == nil {
		return errors.New("Server Session Store not Init")
	}
	session, err := store.Get(r, "aurora_session")
	if session.IsNew {
		return errors.New("Read session from request exception")
	}
	f.User = session.Values["User"].(string)
	f.UUID = session.Values["UUID"].(string)
	return nil
}
