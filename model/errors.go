package model

import "errors"

var (
	ErrNotifierIsRunning = errors.New("notifier is running")
	ErrQueueIsEmpty      = errors.New("queue is empty")
	ErrNotFound          = errors.New("not found")
)
