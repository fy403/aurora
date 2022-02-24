package client_test

import (
	"aurora/client"
	"encoding/json"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	loginURl = "http://localhost/auth"
	tasksUrl = "http://localhost/tasks"
	username = "admin"
	password = "password"
)

func TestSendSyncWithTask(t *testing.T) {
	t.Parallel()
	initTasks()
	connector := client.NewAuroraConnector(loginURl, tasksUrl)
	connector.Init(username, password)
	requestOBJ := &client.CenterRequest{
		TaskType:  "task",
		Timestamp: time.Now().Unix(),
		Signatures: []*client.Signature{
			&addTask0,
		},
		SleepDuration: rand.Intn(1000),
	}
	responseOBJ, err := connector.SendSync(requestOBJ)
	assert.NoError(t, err)
	b, err := json.Marshal(responseOBJ)
	assert.NoError(t, err)
	// (1+1)=?
	t.Logf("responseOBJ: %s", string(b))
}

func TestSendSyncWithGroup(t *testing.T) {
	t.Parallel()
	initTasks()
	connector := client.NewAuroraConnector(loginURl, tasksUrl)
	connector.Init(username, password)
	requestOBJ := &client.CenterRequest{
		TaskType:  "group",
		Timestamp: time.Now().Unix(),
		Signatures: []*client.Signature{
			&addTask0,
			&addTask1,
			&addTask2,
		},
		SleepDuration:   rand.Intn(1000),
		SendConcurrency: 2, // max counts per send subtask
	}
	responseOBJ, err := connector.SendSync(requestOBJ)
	assert.NoError(t, err)
	b, err := json.Marshal(responseOBJ)
	assert.NoError(t, err)
	// (1+1)=? (2+2)=? (5+6)=?
	t.Logf("responseOBJ: %s", string(b))
}

func TestSendSyncWithChain(t *testing.T) {
	t.Parallel()
	initTasks()
	connector := client.NewAuroraConnector(loginURl, tasksUrl)
	connector.Init(username, password)
	requestOBJ := &client.CenterRequest{
		TaskType:  "chain",
		Timestamp: time.Now().Unix(),
		Signatures: []*client.Signature{
			&addTask0,
			&addTask1,
			&addTask2,
			&multiplyTask0,
		},
		SleepDuration: rand.Intn(1000),
	}
	responseOBJ, err := connector.SendSync(requestOBJ)
	assert.NoError(t, err)
	b, err := json.Marshal(responseOBJ)
	assert.NoError(t, err)
	// ((((1 + 1) + (2 + 2)) + (5 + 6)) * 4) = ?
	t.Logf("responseOBJ: %s", string(b))
}

func TestSendSyncWithChord(t *testing.T) {
	t.Parallel()
	initTasks()
	connector := client.NewAuroraConnector(loginURl, tasksUrl)
	connector.Init(username, password)
	requestOBJ := &client.CenterRequest{
		TaskType:  "chord",
		Timestamp: time.Now().Unix(),
		Signatures: []*client.Signature{
			&addTask0,
			&addTask1,
		},
		SleepDuration:   rand.Intn(1000),
		SendConcurrency: 2,
		CallBack:        &multiplyTask1,
	}
	responseOBJ, err := connector.SendSync(requestOBJ)
	assert.NoError(t, err)
	b, err := json.Marshal(responseOBJ)
	assert.NoError(t, err)
	t.Logf("responseOBJ: %s", string(b))
}
