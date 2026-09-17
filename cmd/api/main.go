package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"kvstore/internal/handler"
	"kvstore/internal/repository"
	"kvstore/internal/wal"
)

var WALFilePath = "wal.log"

func main(){
	w, store := Init()

	// replay wal log 
	err := ReApplyWAL(w, store)
	if err != nil {
		log.Fatal(err)
	}

	// enable write ahead logging after the replay
	store.EnableWAL()

	h := handler.NewHandler(store)
	mux := http.NewServeMux()

	/* -- ROUTES -- */
	mux.HandleFunc("GET /", handler.Health)

	mux.HandleFunc("GET /kv/all", h.GETKeys)
	mux.HandleFunc("GET /kv/all/", h.GETKeys)

	mux.HandleFunc("GET /kv/{key}", h.GETvalue)
	mux.HandleFunc("GET /kv/{key}/", h.GETvalue)

	mux.HandleFunc("POST /kv", h.POSTkeyvalue)
	mux.HandleFunc("POST /kv/", h.POSTkeyvalue)

	mux.HandleFunc("POST /kv/clear", h.POSTClear)
	mux.HandleFunc("POST /kv/clear/", h.POSTClear)

	mux.HandleFunc("PUT /kv/{key}", h.PUTvalue)
	mux.HandleFunc("PUT /kv/{key}/", h.PUTvalue)

	mux.HandleFunc("DELETE /kv/{key}", h.DELvalue)
	mux.HandleFunc("DELETE /kv/{key}/", h.DELvalue)

	err = http.ListenAndServe(":8080", mux)
	if errors.Is(err, http.ErrServerClosed){
		fmt.Println("Server Closed")
	} else if err != nil{
		fmt.Printf("error : %s\n", err)
	}
}

func Init() (*wal.WAL, *repository.InMemory) {
	// wal object used by the store service for write-ahead logging
	w, err := wal.Open(WALFilePath)
	if err != nil {
		log.Fatalf("WAL initialization failed; refusing to start: %v", err)
	}	

	store := repository.NewInMemoryStore(w)

	return w, store
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