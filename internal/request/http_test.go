package request_test

import (
	"aurora/internal/request"
	"aurora/internal/tasks"
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	addTask0, addTask1, addTask2                      tasks.Signature
	multiplyTask0, multiplyTask1                      tasks.Signature
	sumIntsTask, sumFloatsTask, concatTask, splitTask tasks.Signature
	panicTask                                         tasks.Signature
	longRunningTask                                   tasks.Signature
)

func TestValidate(t *testing.T) {
	obj := &request.CenterRequest{
		UUID:      "2343543543",
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

func TestDoRequest(t *testing.T) {
	// req, err := http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	initTasks()
	obj := request.CenterRequest{
		UUID:      "2343543543",
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

var initTasks = func() {
	addTask0 = tasks.Signature{
		Name: "add",
		Args: []tasks.Arg{
			{
				Type:  "int64",
				Value: 1,
			},
			{
				Type:  "int64",
				Value: 1,
			},
		},
	}

	addTask1 = tasks.Signature{
		Name: "add",
		Args: []tasks.Arg{
			{
				Type:  "int64",
				Value: 2,
			},
			{
				Type:  "int64",
				Value: 2,
			},
		},
	}

	addTask2 = tasks.Signature{
		Name: "add",
		Args: []tasks.Arg{
			{
				Type:  "int64",
				Value: 5,
			},
			{
				Type:  "int64",
				Value: 6,
			},
		},
	}

	multiplyTask0 = tasks.Signature{
		Name: "multiply",
		Args: []tasks.Arg{
			{
				Type:  "int64",
				Value: 4,
			},
		},
	}

	multiplyTask1 = tasks.Signature{
		Name: "multiply",
	}

	sumIntsTask = tasks.Signature{
		Name: "sum_ints",
		Args: []tasks.Arg{
			{
				Type:  "[]int64",
				Value: []int64{1, 2},
			},
		},
	}

	sumFloatsTask = tasks.Signature{
		Name: "sum_floats",
		Args: []tasks.Arg{
			{
				Type:  "[]float64",
				Value: []float64{1.5, 2.7},
			},
		},
	}

	concatTask = tasks.Signature{
		Name: "concat",
		Args: []tasks.Arg{
			{
				Type:  "[]string",
				Value: []string{"foo", "bar"},
			},
		},
	}

	splitTask = tasks.Signature{
		Name: "split",
		Args: []tasks.Arg{
			{
				Type:  "string",
				Value: "foo",
			},
		},
	}

	panicTask = tasks.Signature{
		Name: "panic_task",
	}

	longRunningTask = tasks.Signature{
		Name: "long_running_task",
	}
}
