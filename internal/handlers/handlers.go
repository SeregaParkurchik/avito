package handlers

import (
	"avito_shop/internal/authentication"
	"avito_shop/internal/core"
	"avito_shop/internal/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type UserHandler struct {
	service core.Interface
}

func NewUserHandler(service core.Interface) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Auth(w http.ResponseWriter, r *http.Request) {
	var newEmployees models.Employee

	if err := json.NewDecoder(r.Body).Decode(&newEmployees); err != nil {
		http.Error(w, "Невалидные данные", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	tokenString, err := h.service.Auth(r.Context(), &newEmployees, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	response := authentication.RegisterResponse{AccessToken: tokenString}
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) BuyItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	item := vars["item"]

	username, ok := r.Context().Value(usernameKey).(string)
	if !ok {
		http.Error(w, "не удалось извлечь имя пользователя из контекста", http.StatusUnauthorized)
		return
	}

	err := h.service.BuyItem(r.Context(), item, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) SendCoin(w http.ResponseWriter, r *http.Request) {
	var newSend models.SendCoin

	if err := json.NewDecoder(r.Body).Decode(&newSend); err != nil {
		http.Error(w, "Невалидные данные", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	username, ok := r.Context().Value(usernameKey).(string)
	if !ok {
		http.Error(w, "не удалось извлечь имя пользователя из контекста", http.StatusUnauthorized)
		return
	}

	err := h.service.SendCoin(r.Context(), &newSend, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) Info(w http.ResponseWriter, r *http.Request) {
	username, ok := r.Context().Value(usernameKey).(string)
	if !ok {
		http.Error(w, "не удалось извлечь имя пользователя из контекста", http.StatusUnauthorized)
		return
	}
	response, err := h.service.Info(r.Context(), username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(jsonResponse)
}
