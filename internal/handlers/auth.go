package handlers

import (
	"errors"
	"net/http"

	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/auth"
	"github.com/PCC-Grupo-11/TB2-Trabajo-final/internal/protocol"
)

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req protocol.RegisterRequest
	if err := protocol.ReadMessage(r.Body, &req); err != nil {
		protocol.WriteJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Username == "" || req.Password == "" {
		protocol.WriteJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "username and password required"})
		return
	}

	if err := h.AuthSvc.Register(r.Context(), req.Username, req.Password); err != nil {
		if errors.Is(err, auth.ErrDuplicateUser) {
			protocol.WriteJSON(w, http.StatusConflict, protocol.ErrorResponse{Error: err.Error()})
		} else {
			protocol.WriteJSON(w, http.StatusInternalServerError, protocol.ErrorResponse{Error: "registration failed"})
		}
		return
	}

	protocol.WriteJSON(w, http.StatusCreated, map[string]string{"message": "user created"})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req protocol.LoginRequest
	if err := protocol.ReadMessage(r.Body, &req); err != nil {
		protocol.WriteJSON(w, http.StatusBadRequest, protocol.ErrorResponse{Error: "invalid request body"})
		return
	}

	token, err := h.AuthSvc.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCreds) {
			protocol.WriteJSON(w, http.StatusUnauthorized, protocol.ErrorResponse{Error: err.Error()})
		} else {
			protocol.WriteJSON(w, http.StatusInternalServerError, protocol.ErrorResponse{Error: "login failed"})
		}
		return
	}

	protocol.WriteJSON(w, http.StatusOK, protocol.AuthResponse{Token: token})
}
