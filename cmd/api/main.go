package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"kvstore/internal/handler"
	"kvstore/internal/repository"
	"kvstore/internal/snapshot"
	"kvstore/internal/wal"
)

var WALFilePath = "wal.log"
var SnapshotFilePath = "snapshot.json"
var wg sync.WaitGroup

func main() {
	/* - - - - INITIALIZATION - - - - */
	w, store, snap := Init(WALFilePath, SnapshotFilePath)

	// replay wal log
	err := ReApplyWAL(w, store, snap)
	if err != nil {
		log.Fatal(err)
	}

	// channel to communicate with ticker routine used to take snapshots
	ctx, cancel := context.WithCancel(context.Background())

	snapshotErr := make(chan error, 1)
	startSnapshotTimer(ctx, snap, snapshotErr)

	// enable Allow services to log to WAL after the replay
	store.EnableWAL()

	h := handler.NewHandler(store)
	mux := http.NewServeMux()

	/* - - - - ROUTES - - - - */
	mux.HandleFunc("GET /", handler.Health)

	mux.HandleFunc("GET /kv", h.GETKeys)
	mux.HandleFunc("GET /kv/", h.GETKeys)

	mux.HandleFunc("GET /kv/{key}", h.GETvalue)
	mux.HandleFunc("GET /kv/{key}/", h.GETvalue)

	mux.HandleFunc("POST /kv", h.POSTkeyvalue)
	mux.HandleFunc("POST /kv/", h.POSTkeyvalue)

	mux.HandleFunc("PUT /kv/{key}", h.PUTvalue)
	mux.HandleFunc("PUT /kv/{key}/", h.PUTvalue)

	mux.HandleFunc("DELETE /kv", h.POSTClear)
	mux.HandleFunc("DELETE  /kv/", h.POSTClear)

	mux.HandleFunc("DELETE /kv/{key}", h.DELvalue)
	mux.HandleFunc("DELETE /kv/{key}/", h.DELvalue)

	err = http.ListenAndServe(":8080", mux)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("Server Closed")
	} else if err != nil {
		fmt.Printf("error : %s\n", err)
	}

	/* - - - - SHUTDOWN - - - - */
	go func() {
		cancel()
		wg.Wait()
	}()
	err = <-snapshotErr
	if err != nil {
		log.Fatal(err)
	}
	// sync os file cache to hard-disk if any on shutdown
	if err = w.Close(); err != nil {
		log.Fatalf("cannot shutdown wal; error - %v", err)
	}
}

func Init(WALFilePath string, SnapshotFilePath string) (*wal.WAL, *repository.InMemory, *snapshot.Snap) {
	// wal object used by the store service for write-ahead logging
	w, err := wal.Open(WALFilePath)
	if err != nil {
		log.Fatalf("WAL initialization failed; refusing to start: %v", err)
	}

	store := repository.NewInMemoryStore(w)

	snap := snapshot.NewSnapshotLogger(w, store, SnapshotFilePath)
	err = snap.InitSnap()
	if err != nil {
		log.Fatalf("snapshot initialization failed; refusing to start : %v", err)
	}

	return w, store, snap
}

func ReApplyWAL(w *wal.WAL, store *repository.InMemory, snap *snapshot.Snap) error {
	latestSnap, err := snap.GetLatestSnapshot()
	if err != nil {
		return fmt.Errorf("cannot get latest snapshot : %v", err)
	}

	// replay the wal records after the store is created
	records, err := w.Replay(latestSnap.WalCheckpoint)

	if err != nil {
		return fmt.Errorf("WAL replay (parsing log file) failed; refusing to start %v", err)
	}

	// set latest state
	err = store.SetState(latestSnap.State)
	if err != nil {
		return fmt.Errorf("cannot set last state : %v", err)
	}

	err = store.Apply(records)
	if err != nil {
		return fmt.Errorf("WAL replay (applying records) failed; refusing to start %v", err)
	}
	return nil
}

// startSnapshotTimer takes a snapshot every 10 seconds until the context ends.
func startSnapshotTimer(ctx context.Context, snap *snapshot.Snap, snapshotErr chan<- error) {
	wg.Add(1)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := snap.TakeSnapshot()
				if err != nil {
					snapshotErr <- fmt.Errorf("snapshot logging failed; refusing to continue : %v", err)
					wg.Done()
					return
				}
			case <-ctx.Done():
				snapshotErr <- nil
				wg.Done()
				return
			}
		}
	}()
}
