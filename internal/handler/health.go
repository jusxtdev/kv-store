package handler

import "net/http"

func Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
	}
	WriteJSONResponse(w, http.StatusOK, Response{true, "alive", nil})
}
