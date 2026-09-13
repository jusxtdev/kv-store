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

	mux.HandleFunc("GET /key", h.GETvalue)

	err := http.ListenAndServe(":8080", mux)
	if errors.Is(err, http.ErrServerClosed){
		fmt.Println("Server Closed")
	} else if err != nil{
		fmt.Printf("error : %s\n", err)
	}
}