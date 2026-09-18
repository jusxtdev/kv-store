/*
	TODO

* modify the wal record to have ids for each operation
* FIX - make the types consistent accross memory.go and record.go
* implement getCurrState function in memory.go
- find a way to fire state logging function periodically or on POST clear operation
* write init state logger function
* setup state logger
  - getCurrState()
  - get current operation index (walCheckpoint)
  - construct stateRecord
  - append stateRecord to state.json

- modify wal replay
  - fetch latest state
  - extract walCheckpoint
  - open wal.log and iterate till you get the walCheckpoint
  - and execute operations *after* that checkpoint
*/

package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"kvstore/internal/repository"
	"kvstore/internal/wal"
	"os"
)

type Snap struct {
	snapshotFilePath string
	wal              *wal.WAL
	store            *repository.InMemory
}

type SnapRecord struct {
	WalCheckpoint int              `json:"wal_checkpoint"`
	State         repository.KVMap `json:"data"`
}

func NewSnapshotLogger(wal *wal.WAL, store *repository.InMemory, path string) *Snap {
	/*
	  return a pointer to Snap object to interat with SnapShot logger
	*/
	snap := &Snap{
		wal:              wal,
		store:            store,
		snapshotFilePath: path,
	}
	return snap
}

func (snap *Snap) InitSnap() error {
	/*
	  Initialize the snap log file
	*/
	var file *os.File
	var err error

	_, err = os.Stat(snap.snapshotFilePath)
	if os.IsNotExist(err) {
		file, err = os.Create(snap.snapshotFilePath)
		if err != nil {
			return errors.New("cannot create snapshot file")
		}
	}
	defer file.Close()

	return nil
}

func (snap *Snap) TakeSnapshot() error {
	state := snap.store.GetCurrentState()
	walCheckpoint := snap.wal.GetWALCheckpoint()
	record := SnapRecord{
		WalCheckpoint: walCheckpoint,
		State:         state,
	}

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("cannot marshal state record : %v", err)
	}

	err = os.WriteFile(snap.snapshotFilePath, data, 0644)
	if err != nil {
		return fmt.Errorf("cannot write snapshot : %v", err)
	}

	return nil
}

func (snap *Snap) GetLatestSnapshot() (SnapRecord, error) {
	// return if snapshot file is empty
	if walCheckPoint := snap.wal.GetWALCheckpoint(); walCheckPoint == 0 {
		return SnapRecord{
			WalCheckpoint: 0,
			State:         make(repository.KVMap),
		}, nil
	}

	data, err := os.ReadFile(snap.snapshotFilePath)
	if err != nil {
		return SnapRecord{}, fmt.Errorf("cannot read snapshot file : %v", err)
	}

	var record SnapRecord

	err = json.Unmarshal(data, &record)
	if err != nil {
		return SnapRecord{}, fmt.Errorf("cannot unmarshal snapshot file : %v", err)
	}
	return record, nil
}
