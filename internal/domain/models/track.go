package models

import "time"

type Track struct {
	ID            int64
	Title         string
	CreatedAt     time.Time
	OriginBucvket string
	OriginKey     string
	HLSBucket     *string
	HLSPrefix     *string
}
