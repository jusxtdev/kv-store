package wal

import (
	"errors"
	"fmt"
	"os"
)

type WAL struct {
	logFilePath string
}

func Open(logFilePath string) (*WAL, error) {
	var file *os.File
	var err error

	// if file doesn't exist
	_, err = os.Stat(logFilePath)
	if os.IsNotExist(err){
		file, err = os.Create(logFilePath)
		if err != nil {
			return nil, errors.New("cannot create wal file")
		}
	}
	defer file.Close()

	// return wal
	return &WAL{logFilePath}, nil
}

func (w *WAL) Serialize(operation, key, value string) []byte {
	r := fmt.Sprintf("%s %s %s\n", operation, key, value)
	return []byte(r)
}