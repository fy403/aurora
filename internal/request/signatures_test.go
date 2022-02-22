package request_test

import "aurora/internal/tasks"

var (
	addTask0, addTask1, addTask2                      tasks.Signature
	multiplyTask0, multiplyTask1                      tasks.Signature
	sumIntsTask, sumFloatsTask, concatTask, splitTask tasks.Signature
	panicTask                                         tasks.Signature
	longRunningTask                                   tasks.Signature
)
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
