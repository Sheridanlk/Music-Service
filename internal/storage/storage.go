package storage

import (
	"errors"
)

type ByteRange struct {
	Start int64
	End   int64
}

var (
	ErrTrackNotFound = errors.New("track not found")
)

const (
	StatusReady      = "ready"
	StatusError      = "error"
	StatusProcessing = "processing"
	StatusPending    = "pending"
)
