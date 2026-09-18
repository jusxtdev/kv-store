package wal

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type WAL struct {
	logFilePath string
	currentOprId int	// 1 based index for each record
}

// '|' is used as delimeter in Record strings (delimeted in Serialize()) 

func (wal *WAL) SetOperationId() error {
	var lastline string
	// first get the last line in the wal log
	file, err := os.Open(wal.logFilePath)
	if err != nil {
		return fmt.Errorf("cannot open wal log file - %s", wal.logFilePath)
	}

	// iterate on each line
	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		lastline = scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error occurred while reading log file %s", err)
	}

	if len(lastline) == 0 {
		wal.currentOprId = 0
		return nil
	}
	rec, err := parseLine(lastline)
	if err != nil {
		return fmt.Errorf("error occurred while reading log file %s", err)
	}
	wal.currentOprId = rec.OpId + 1
	return nil
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

	wal := WAL{logFilePath: logFilePath}
	err = wal.SetOperationId()
	if err != nil {
		return nil, err
	}
	// return wal
	return &wal, nil
}

func (w *WAL) GetWALCheckpoint() int {
	/*
	Used to get the operation id of latest operation
	*/
	return w.currentOprId
}

func (w *WAL) Serialize(operation, key, value string) []byte {
	r := fmt.Sprintf("%d|%s|%s|%s\n", w.currentOprId, operation, key, value)
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

	// increment wal operation id
	w.currentOprId = w.currentOprId + 1
	
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

func (w *WAL) Close() error {
	// open log file
	f, err := os.OpenFile(
		w.logFilePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return fmt.Errorf("cannot open wal log file - %s", w.logFilePath)
	}
	defer f.Close()

	// ask os to write the file cache to hard-disk before closing
	if err := f.Sync(); err != nil {
		return fmt.Errorf("cannot sync wal log file - %s", w.logFilePath)
	}
	return nil
}

func (w *WAL) Replay(checkpoint int) ([]Record, error) {
	var records []Record

	// open the wal logfile
	file, err := os.Open(w.logFilePath)
	if err != nil {
		return []Record{}, fmt.Errorf("cannot open wal log file - %s", w.logFilePath)
	}

	// iterate on each line
	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		line := scanner.Text()
		
		rec, err := parseLine(line)
		if err != nil {
			// skip current line if an error occurred
			// might add the error info to a log file if there is one
			continue
		}
		if rec.OpId < checkpoint{
			// skip the record before and till the checkpoint
			continue
		}

		records = append(records, rec)
	}
	if err := scanner.Err(); err != nil {
		return []Record{}, fmt.Errorf("error occurred while reading log file %s", err)
	}

	// return the records
	return records, nil
}

func parseLine(line string) (Record, error) {
	if len(line) == 0 {
		// ignore empty lines
		return Record{}, fmt.Errorf("empty line")
	}

	var rec Record
	lineSlice := strings.Split(line, "|")
	// if the line is like - "hello there" i.e. neither a valid delimeter nor a valid operation written
	// then an empty record with error is returned


	// parse operation
	var operation Operation
	oprStr := lineSlice[1]
	switch oprStr {
	case "SET":
		operation = Set
	case "UPDATE":
		operation = Update
	case "DELETE":
		operation = Delete
	case "CLEAR":
		operation = Clear
	default:
		// as there is not a valid operation, just return a Record with -1 as the Operation value
		return Record{}, fmt.Errorf("invalid operation")
	}
	rec.Operation = operation

	// parse operation id
	opIdStr := lineSlice[0]
	opId, err := strconv.Atoi(opIdStr)
	if err != nil {
		return Record{}, fmt.Errorf("some error while parsing operation id : %v", err)
	}
	rec.OpId = opId

	// opId Opr key val
	if len(lineSlice) >= 3 {
		rec.Key = string(lineSlice[2])
	}
	if len(lineSlice) >= 4 {
		rec.Value = string(lineSlice[3])
	}
	return rec, nil
}