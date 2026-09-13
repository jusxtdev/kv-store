package main

import (
	"errors"
	"fmt"
	"net/http"

	"kvstore/internal/handler"
	"kvstore/internal/repository"
)

func main(){
	store := repository.NewInMemoryStore()
	h := handler.NewHandler(store)
	
	mux := http.NewServeMux()

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