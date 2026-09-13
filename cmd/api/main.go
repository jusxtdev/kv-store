package main

import (
	"errors"
	"fmt"
	"kvstore/internal/handler"
	"net/http"
)

func main(){
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handler.Health)

	err := http.ListenAndServe(":8080", mux)
	if errors.Is(err, http.ErrServerClosed){
		fmt.Println("Server Closed")
	} else if err != nil{
		fmt.Printf("error : %s\n", err)
	}
}