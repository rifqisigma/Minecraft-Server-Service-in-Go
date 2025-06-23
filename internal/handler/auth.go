package handler

import (
	"encoding/json"
	"minecrat_go/dto"
	"minecrat_go/helper/middleware"
	"minecrat_go/helper/utils"
	"minecrat_go/internal/usecase"
	"net/http"
)

type AuthHandler struct {
	authUC usecase.AuthUseCase
}

func NewAuthHandler(authUC usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUC}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input dto.Register

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.Email == "" || input.Username == "" || input.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if err := h.authUC.Register(&input); err != nil {
		switch err {
		case utils.ErrInvalidEmail:
			utils.WriteError(w, http.StatusBadRequest, "invalid type email")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "link verification has been send or your email",
	})

}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input dto.Login
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.Email == "" || input.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}
	token, err := h.authUC.Login(&input)
	if err != nil {
		switch err {
		case utils.ErrInvalidEmail:
			utils.WriteError(w, http.StatusBadRequest, "invalid type email")
			return
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"token_jwt": token,
	})

}

func (h *AuthHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.AuthKey).(*utils.JWTClaims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "")
		return
	}

	if err := h.authUC.DeleteUser(claims.UserID); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "succed delete this account",
	})
}
