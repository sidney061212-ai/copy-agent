package service

import (
	"os"
	"path/filepath"
	"sync"
)

type RotatingWriter struct {
	mu      sync.Mutex
	file    *os.File
	path    string
	maxSize int64
	curSize int64
}

func NewRotatingWriter(path string, maxSize int64) (*RotatingWriter, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return &RotatingWriter{
		file:    file,
		path:    path,
		maxSize: maxSize,
		curSize: info.Size(),
	}, nil
}

func (writer *RotatingWriter) Write(data []byte) (int, error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	if writer.file == nil {
		return 0, os.ErrClosed
	}

	n, err := writer.file.Write(data)
	writer.curSize += int64(n)
	if writer.curSize > writer.maxSize {
		if rotateErr := writer.rotate(); rotateErr != nil && err == nil {
			err = rotateErr
		}
	}
	return n, err
}

func (writer *RotatingWriter) rotate() error {
	if writer.file != nil {
		if err := writer.file.Close(); err != nil {
			return err
		}
	}

	backupPath := writer.path + ".1"
	_ = os.Remove(backupPath)
	if err := os.Rename(writer.path, backupPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	file, err := os.OpenFile(writer.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		writer.file = nil
		writer.curSize = 0
		return err
	}

	writer.file = file
	writer.curSize = 0
	return nil
}

func (writer *RotatingWriter) Close() error {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if writer.file == nil {
		return nil
	}
	err := writer.file.Close()
	writer.file = nil
	return err
}
