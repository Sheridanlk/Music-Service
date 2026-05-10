package media

import (
	"io"
	"os"
)

// WriteToFile writes the content from the reader to a file at the specified path.
func WriteToFile(path string, reader io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	return nil
}

// OpenFile opens the file at the given path and returns the file handle and its size.
func OpenFile(path string) (*os.File, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}

	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, 0, err
	}

	return f, stat.Size(), nil
}
