package media

import "fmt"

func GenerateTrackOriginKey(id int64, ext string) string {
	return fmt.Sprintf("tracks/%d/source/original%s", id, ext)
}

func GenerateTrackHLSKey(id int64) string {
	return fmt.Sprintf("tracks/%d/hls/aac_128/", id)
}
