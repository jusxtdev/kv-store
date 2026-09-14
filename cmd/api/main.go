package main

import (
	"errors"
	"fmt"
	"net/http"

	"kvstore/internal/handler"
	"kvstore/internal/repository"
	"kvstore/internal/wal"
)

func main(){
	// wal object used by the store service for write-ahead logging
	w := wal.NewWAL("wal.log")

	store := repository.NewInMemoryStore(w)

	// replay the wal records after the store is created

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

	err := http.ListenAndServe(":8080", mux)
	if errors.Is(err, http.ErrServerClosed){
		fmt.Println("Server Closed")
	} else if err != nil{
		fmt.Printf("error : %s\n", err)
	}
}