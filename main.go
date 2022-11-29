package main

import (
	"github.com/joho/godotenv"
	"github.com/katakeda/lantrn-api-go/app"
)

func main() {
	godotenv.Load()

	app := app.App{}
	app.Initialize()
	app.Run()
}
