package handler

import (
	"encoding/json"
	"minecrat_go/dto"
	"minecrat_go/helper/middleware"
	"minecrat_go/helper/utils"
	"minecrat_go/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type BedrockHandler struct {
	bduc usecase.BedrockUC
}

func NewBedrockHandler(bduc usecase.BedrockUC) *BedrockHandler {
	return &BedrockHandler{bduc}
}

func (h *BedrockHandler) CreateWorld(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.AuthKey)
	claims, ok := claimsRaw.(*utils.JWTClaims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "invalid jwt")
		return
	}

	var req dto.ServerParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := utils.ValidateReq(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	req.Creator = claims.UserID
	if err := h.bduc.CreateServer(&req); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *BedrockHandler) DeleteWorld(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.AuthKey)
	claims, ok := claimsRaw.(*utils.JWTClaims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "invalid jwt")
		return
	}

	params := mux.Vars(r)
	paramsWorld := params["world"]

	if err := h.bduc.DeleteWorld(claims.UserID, paramsWorld); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *BedrockHandler) StartWorld(w http.ResponseWriter, r *http.Request) {
	var req dto.StartServerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.bduc.StartServer(&req); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *BedrockHandler) StopWorld(w http.ResponseWriter, r *http.Request) {
	var req dto.StartServerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	params := mux.Vars(r)
	paramsWorld := params["world"]

	if err := h.bduc.StopServer(paramsWorld); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *BedrockHandler) EditWorld(w http.ResponseWriter, r *http.Request) {
	claimsRaw := r.Context().Value(middleware.AuthKey)
	claims, ok := claimsRaw.(*utils.JWTClaims)
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "invalid jwt")
		return
	}

	params := mux.Vars(r)
	paramsId, err := strconv.ParseUint(params["id"], 10, 64)
	paramsWorld := params["world"]
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req dto.ServerParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := utils.ValidateReq(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	req.Creator = claims.UserID
	if err := h.bduc.EditWorld(&req, uint(paramsId), paramsWorld); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *BedrockHandler) SendCommand(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CMD string `json:"cmd"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	params := mux.Vars(r)
	paramsWorld := params["world"]

	if err := h.bduc.SendCommandforAPI(paramsWorld, req.CMD); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *BedrockHandler) BanPlayer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	paramsName := params["name"]
	paramsWorld := params["world"]

	if err := h.bduc.BanPlayer(paramsWorld, paramsName); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *BedrockHandler) KickPlayer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	paramsName := params["name"]
	paramsWorld := params["world"]

	if err := h.bduc.KickPlayer(paramsWorld, paramsName); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *BedrockHandler) GetPermissionPlayer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	paramsWorld := params["world"]

	response, err := h.bduc.GetPermissionPlayer(paramsWorld)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *BedrockHandler) GetWorlds(w http.ResponseWriter, r *http.Request) {
	response, err := h.bduc.GetWorlds()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *BedrockHandler) GetWorldAndPlayers(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	paramsWorld := params["world"]

	response, err := h.bduc.GetWorldAndPlayers(paramsWorld)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

func (h *BedrockHandler) CreateOrUpdatePermissionPlayer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	paramsWorld := params["world"]

	var req dto.PermissionPlayer
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.bduc.CreateOrUpdatePermissions(&req, paramsWorld); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *BedrockHandler) DeletePermissionPlayer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	paramsWorld := params["world"]
	paramsUid := params["xuid"]

	if err := h.bduc.DeletePermission(paramsUid, paramsWorld); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *BedrockHandler) GetLogsServer(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	paramsWorld := params["world"]

	logs, err := h.bduc.GetServerLogs(paramsWorld)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, logs)
}

func (h *BedrockHandler) CreatePriority(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	paramsWorld := params["world"]

	var req dto.Allowlist
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.bduc.CreatePriority(&req, paramsWorld); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *BedrockHandler) DeletePriority(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	paramsWorld := params["world"]
	paramsXuid := params["xuid"]

	if err := h.bduc.DeletePriority(paramsXuid, paramsWorld); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, nil)
}

func (h *BedrockHandler) GetPriority(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	paramsWorld := params["world"]

	response, err := h.bduc.GetPriority(paramsWorld)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, response)
}
