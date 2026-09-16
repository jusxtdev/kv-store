package wal

type Operation int

const (
	Set Operation = iota
	Update
	Delete
	Clear
)

type Record struct{
	Operation Operation
	Key string
	Value string
}