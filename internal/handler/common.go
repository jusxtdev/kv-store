package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Status bool `json:"status"`
	Msg string `json:"msg"`
	Data any `json:"data"`
}

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data Response){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			fmt.Println(err)
	}
}