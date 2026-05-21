package main

import (
	"greentrust-hackathon/internal/handler/rest"
	"greentrust-hackathon/internal/repository"
	"greentrust-hackathon/internal/service"
	"greentrust-hackathon/pkg/bcrypt"
	"greentrust-hackathon/pkg/config"
	"greentrust-hackathon/pkg/database/mariadb"
	"greentrust-hackathon/pkg/jwt"
	"greentrust-hackathon/pkg/middleware"
	"log"
)

func main() {
	config.LoadEnvironment()

	db, err := mariadb.ConnectDatabase()
	if err != nil {
		log.Fatal(err)
	}

	err = mariadb.Migrate(db)
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewRepository(db)
	bcrypt := bcrypt.Init()
	jwt := jwt.Init()
	svc := service.NewService(repo, bcrypt, jwt)

	middleware := middleware.Init(svc, jwt)
	r := rest.NewRest(svc, middleware)
	r.MountEndpoint()

	r.Run()
}
