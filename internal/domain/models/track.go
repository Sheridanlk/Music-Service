package models

import "time"

type Track struct {
	ID           int64
	Title        string
	CreatedAt    time.Time
	OriginBucket string
	OriginKey    string
	HLSBucket    *string
	HLSPrefix    *string
}

type TrackListItem struct {
	ID        int64
	Title     string
	CreatedAt time.Time
}
