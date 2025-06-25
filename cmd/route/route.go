package route

import (
	"minecrat_go/helper/middleware"
	"minecrat_go/internal/handler"
	"net/http"

	"github.com/gorilla/mux"
)

func SetupRoute(authHandler *handler.AuthHandler, bedrockHandler *handler.BedrockHandler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
	r.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)

	userRoute := r.PathPrefix("/user").Subrouter()
	userRoute.Use(middleware.AuthMiddeware)

	userRoute.HandleFunc("/delete", authHandler.DeleteUser).Methods(http.MethodDelete)

	bedrockRoute := r.PathPrefix("/bedrock").Subrouter()
	bedrockRoute.Use(middleware.AuthMiddeware)

	bedrockRoute.HandleFunc("/create", bedrockHandler.CreateWorld).Methods(http.MethodPost)
	bedrockRoute.HandleFunc("/{world}/delete", bedrockHandler.DeleteWorld).Methods(http.MethodDelete)
	bedrockRoute.HandleFunc("/{world}/{id}/update", bedrockHandler.EditWorld).Methods(http.MethodPut)
	bedrockRoute.HandleFunc("/start", bedrockHandler.StartWorld).Methods(http.MethodPost)
	bedrockRoute.HandleFunc("/{world}/stop", bedrockHandler.StopWorld).Methods(http.MethodPost)
	bedrockRoute.HandleFunc("/{world}/command", bedrockHandler.SendCommand).Methods(http.MethodPost)
	bedrockRoute.HandleFunc("/{world}/command/ban/{name}", bedrockHandler.BanPlayer).Methods(http.MethodPost)
	bedrockRoute.HandleFunc("/{world}/command/kick/{name}", bedrockHandler.KickPlayer).Methods(http.MethodPost)
	bedrockRoute.HandleFunc("/{world}/get-permission-players", bedrockHandler.GetPermissionPlayer).Methods(http.MethodGet)
	bedrockRoute.HandleFunc("/get-worlds", bedrockHandler.GetWorlds).Methods(http.MethodGet)
	bedrockRoute.HandleFunc("/{world}/get-world-players", bedrockHandler.GetWorldAndPlayers).Methods(http.MethodGet)
	bedrockRoute.HandleFunc("/{world}/create-or-update-permission", bedrockHandler.CreateOrUpdatePermissionPlayer).Methods(http.MethodPost)
	bedrockRoute.HandleFunc("/{world}/delete-permission/{xuid}", bedrockHandler.DeletePermissionPlayer).Methods(http.MethodDelete)
	bedrockRoute.HandleFunc("/{world}/logs", bedrockHandler.GetLogsServer).Methods(http.MethodGet)
	bedrockRoute.HandleFunc("/{world}/create-priority", bedrockHandler.CreatePriority).Methods(http.MethodPost)
	bedrockRoute.HandleFunc("/{world}/delete-priority/{xuid}", bedrockHandler.DeletePriority).Methods(http.MethodDelete)
	bedrockRoute.HandleFunc("/{world}/get-priority/", bedrockHandler.GetPriority).Methods(http.MethodGet)

	return r
}
