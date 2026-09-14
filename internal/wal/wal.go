package wal

type WAL struct {
	logFilePath string
}

func NewWAL(logFilePath string) *WAL{
	// do the file creation stuff here
	return &WAL{logFilePath}
}

func (w *WAL)Open(){}