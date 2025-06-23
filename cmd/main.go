package main

import (
	"log"
	"minecrat_go/cmd/database"
	"minecrat_go/cmd/route"
	"minecrat_go/internal/handler"
	"minecrat_go/internal/repository"
	"minecrat_go/internal/usecase"
	"net/http"
	"os"
)

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	authRepo := repository.NewAuthRepository(db)
	authUc := usecase.NewAuthUseCase(authRepo)
	authHandler := handler.NewAuthHandler(authUc)

	bedrockRepo := repository.NewBedrockRepo(db)
	bedrockUC := usecase.NewBedrockUC(bedrockRepo)
	bedrockHandler := handler.NewBedrockHandler(bedrockUC)

	r := route.SetupRoute(authHandler, bedrockHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server berjaalan pada port :%s", port)
	http.ListenAndServe(":"+port, r)

}
