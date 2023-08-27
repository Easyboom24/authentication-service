package main

import (
	"go-test/internal/app"
	"go-test/internal/config"
)



func main() {
	cfg := config.GetConfig()
	app.Run(*cfg)	
}
