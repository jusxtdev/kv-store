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

func (w *WAL) Write(record []byte) error {
	// open the file in append mode
	f, err := os.OpenFile(
		w.logFilePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return fmt.Errorf("cannot open wal log file - %s", w.logFilePath)
	}
	defer f.Close()

	// write bytes to the operating systems file cache
	if _, err := f.Write(record); err != nil {
		return fmt.Errorf("cannot write to wal log file - %s, record - %s", w.logFilePath, string(record))
	}

	// ask os to write the record from the os file cache to the hard-disk
	if err := f.Sync(); err != nil {
		return fmt.Errorf("cannot sync wal log file - %s, record - %s", w.logFilePath, string(record))
	}
	return nil
}