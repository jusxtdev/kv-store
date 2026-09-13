package handler

import (
	"kvstore/internal/repository"
	"net/http"
)

type Handler struct {
	store *repository.StoreRepository
}

func NewHandler(store repository.StoreRepository) *Handler {
	return &Handler{
		store: &store,
	}
} 

func (h *Handler) GETvalue(w http.ResponseWriter, r *http.Request){}