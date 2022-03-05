package request

import (
	"aurora/internal/auth"
	"aurora/internal/tasks"
	"errors"
	"net/http"

	validator "github.com/go-playground/validator/v10"
)

type CenterRequest struct {
	UUID            string             `json:"UUID" validate:"max=50"`    // user id
	User            string             `json:"User" validate:"max=15"`    // user name
	BatchID         string             `json:"BatchID" validate:"max=50"` // unique id for request
	Timestamp       int64              `json:"Timestamp" validate:"required"`
	TaskType        string             `json:"TaskType" validate:"required,oneof='task' 'group' 'chord' 'chain'"`
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
		return errors.New("Read session from request exceptly")
	}
	f.User = session.Values["User"].(string)
	f.UUID = session.Values["UUID"].(string)
	return nil
}
