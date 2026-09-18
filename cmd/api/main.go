package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"kvstore/internal/handler"
	"kvstore/internal/repository"
	"kvstore/internal/snapshot"
	"kvstore/internal/wal"
)

var WALFilePath = "wal.log"
var SnapshotFilePath = "snapshot.json"

func main(){
	w, store, snap := Init(WALFilePath, SnapshotFilePath)

	// replay wal log 
	err := ReApplyWAL(w, store)
	if err != nil {
		log.Fatal(err)
	}

	// channel to communicate with ticker routine used by SnapLogger
	ctx, cancel := context.WithCancel(context.Background())

	go func(){
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := snap.TakeSnapshot()
				if err != nil {
					log.Fatalf("snapshot logging failed; refusing to continue : %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// enable Allow services to log to WAL after the replay
	store.EnableWAL()

	h := handler.NewHandler(store)
	mux := http.NewServeMux()

	/* -- ROUTES -- */
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
	if errors.Is(err, http.ErrServerClosed){
		fmt.Println("Server Closed")
	} else if err != nil{
		fmt.Printf("error : %s\n", err)
	}

	// sync os file cache to hard-disk if any on shutdown
	if err = w.Close(); err != nil {
		log.Fatalf("cannot shutdown wal; error - %v", err)
	}
	cancel()
}

func Init(WALFilePath string, SnapshotFilePath string) (*wal.WAL, *repository.InMemory, *snapshot.Snap) {
	// wal object used by the store service for write-ahead logging
	w, err := wal.Open(WALFilePath)
	if err != nil {
		log.Fatalf("WAL initialization failed; refusing to start: %v", err)
	}	

	store := repository.NewInMemoryStore(w)

	snap := snapshot.NewSnapshotLogger(w, store, SnapshotFilePath)

	return w, store, snap
}

func ReApplyWAL(w *wal.WAL, store *repository.InMemory) error {
	// replay the wal records after the store is created
	records, err := w.Replay()
	if err != nil {
		return fmt.Errorf("WAL replay (parsing log file) failed; refusing to start %v", err)
	}

	err = store.Apply(records)
	if err != nil {
		return fmt.Errorf("WAL replay (applying records) failed; refusing to start %v", err)
	}
	return nil
}
