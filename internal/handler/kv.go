package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"kvstore/internal/repository"
)

type Handler struct {
	store repository.StoreRepository
}

func NewHandler(store repository.StoreRepository) *Handler {
	return &Handler{
		store: store,
	}
}

type POSTRequestBody struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PUTRequestBody struct {
	Value string `json:"value"`
}

// Read handlers.

func (h *Handler) GETKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	allkeys := h.store.Keys()
	WriteJSONResponse(w, http.StatusOK, Response{true, "all keys", allkeys})
}

func (h *Handler) GETvalue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	key := r.PathValue("key")

	val, err := h.store.Get(key)
	if err != nil {
		WriteJSONResponse(w, http.StatusNotFound, Response{false, err.Error(), nil})
		return
	}

	WriteJSONResponse(w, http.StatusOK, Response{true, "value for key found", val})
}

func (h *Handler) GETExists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	key := r.PathValue("key")
	exists := h.store.Exists(key)
	if exists {
		WriteJSONResponse(w, http.StatusOK, Response{true, "key exists", nil})
		return
	}
	WriteJSONResponse(w, http.StatusNotFound, Response{false, "key not found", nil})
}

// Write handlers.

func (h *Handler) POSTkeyvalue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var body POSTRequestBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&body)
	if err != nil {
		WriteJSONResponse(w, http.StatusBadRequest, Response{false, "invalid json", nil})
		return
	}

	err = h.store.Set(body.Key, body.Value)
	if err != nil {
		WriteJSONResponse(w, http.StatusConflict, Response{false, err.Error(), nil})
		return
	}

	WriteJSONResponse(w, http.StatusCreated, Response{true, "key added successfully", nil})
}

func (h *Handler) PUTvalue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeMethodNotAllowed(w)
		return
	}

	key := r.PathValue("key")

	var body PUTRequestBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&body)
	if err != nil {
		WriteJSONResponse(w, http.StatusBadRequest, Response{false, "invalid json", nil})
		return
	}

	err = h.store.Update(key, body.Value)
	if err != nil {
		WriteJSONResponse(w, http.StatusNotFound, Response{false, err.Error(), nil})
		return
	}

	WriteJSONResponse(w, http.StatusOK, Response{true, "updated key successfully", nil})
}

func (h *Handler) DELvalue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeMethodNotAllowed(w)
		return
	}

	key := r.PathValue("key")

	err := h.store.Delete(key)
	if err != nil {
		WriteJSONResponse(w, http.StatusNotFound, Response{false, err.Error(), nil})
		return
	}

	WriteJSONResponse(w, http.StatusOK, Response{true, "deleted key successfully", nil})
}

func (h *Handler) POSTClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	if err := h.store.Clear(); err != nil {
		fmt.Println(err)
		WriteJSONResponse(w, http.StatusInternalServerError, Response{false, "internal server error", nil})
		return
	}
	WriteJSONResponse(w, http.StatusOK, Response{true, "cleared keys", nil})

}
