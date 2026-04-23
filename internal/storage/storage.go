package storage

var (
	ErrNotFoundTrackWithStatus = error.New("track not found or not in uploading status")

type ByteRange struct {
	Start int64
	End   int64
}
