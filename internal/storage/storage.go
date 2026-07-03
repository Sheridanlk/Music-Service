package storage

type ByteRange struct {
	Start int64
	End   int64
}

const (
	StatusReady      = "ready"
	StatusError      = "error"
	StatusProcessing = "processing"
	StatusPending    = "pending"
)
