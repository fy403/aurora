package iface

import (
	"aurora/internal/tasks"
)

// Backend - a common interface for all result backends
type Backend interface {
	// Group related functions
	InitGroup(groupUUID string, taskUUIDs []string) error
	GroupCompleted(groupUUID string) (bool, error)
	GroupTaskStates(groupUUID string) ([]*tasks.TaskState, error)
	TriggerChord(groupUUID string) (bool, error)

	// Graph related functions
	InitGraph(graph *tasks.Graph) error
	GraphCompleted(graphUUID string) (bool, error)
	GraphStates(graphUUID string) (*tasks.Graph, error)
	UpdateGraphStates(graph *tasks.Graph) error

	// Setting / getting task state
	SetStatePending(signature *tasks.Signature) error
	SetStateReceived(signature *tasks.Signature) error
	SetStateStarted(signature *tasks.Signature) error
	SetStateRetry(signature *tasks.Signature) error
	SetStateSuccess(signature *tasks.Signature, results []*tasks.TaskResult) error
	SetStateFailure(signature *tasks.Signature, err string) error
	GetState(taskUUID string) (*tasks.TaskState, error)

	// Purging stored stored tasks states and group meta data
	IsAMQP() bool
	PurgeState(taskUUID string) error
	PurgeGroupMeta(groupUUID string) error

	TestConnect() error
}
