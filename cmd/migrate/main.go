package main

import (
	"fmt"
	"log"
	"minecrat_go/cmd/database"
	"minecrat_go/model"
)

func main() {
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("konek db err :%s", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.WorldServer{}, &model.MemberRole{}); err != nil {
		log.Fatalf("migrate dbe rr :%s", err)
	}

	fmt.Println("migrate db berhasil")
}
