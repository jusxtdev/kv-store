package handler

import "net/http"

func Health(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		WriteJSONResponse(w, http.StatusMethodNotAllowed, Response{false, "method not allowed", nil})
	}
	WriteJSONResponse(w, http.StatusOK, Response{true, "alive", nil})
}