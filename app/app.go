package app

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/katakeda/lantrn-api-go/repositories"
	"github.com/katakeda/lantrn-api-go/services"
)

type App struct {
	router *gin.Engine
}

func (app *App) Initialize() {
	db, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln("Failed to connect with DB", err)
	}

	repo, err := repositories.NewRepository(db)
	if err != nil {
		log.Fatalln("Failed to initialize repository", err)
	}

	svc, err := services.NewService(repo)
	if err != nil {
		log.Fatalln("Failed to initialize service", err)
	}

	app.router = gin.Default()
	app.router.GET("/facilities", svc.GetFacilities)
	app.router.GET("/facilities/:id", svc.GetFacility)
}

func (app *App) Run() {
	err := app.router.Run(":" + os.Getenv("PORT"))
	if err != nil {
		log.Fatalln("Failed to run app", err)
	}
}
