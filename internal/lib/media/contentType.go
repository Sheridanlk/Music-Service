package media

import (
	"mime"
	"path/filepath"
	"strings"
)

func DetectContentType(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return "application/octet-stream"
	}

	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}

	switch ext {
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".ts":
		return "video/MP2T"
	case ".aac":
		return "audio/aac"
	default:
		return "application/octet-stream"
	}
}
