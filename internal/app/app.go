package app

import (
	"context"
	"fmt"
	"go-test/internal/config"
	"go-test/internal/handlers"
	"go-test/internal/repository"
	"go-test/pkg/client/mongodb"
	"go-test/pkg/logging"
	"net"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)


func Run(cfg config.Config) {
	logger := logging.GetLogger()
	logger.Info("Create router")
	router := httprouter.New()

	cfgMongo := cfg.MongoDB
	mongoDBClient, err := mongodb.NewClient(context.Background(), cfgMongo.Host, cfgMongo.Port, cfgMongo.Username,
		cfgMongo.Password, cfgMongo.Database, cfgMongo.Auth_db)
	if err != nil {
		panic(err)
	}
	logger.Info("Create storages")
	Storages := repository.NewStorages(mongoDBClient,*logger)

	logger.Info("Register user handler")
	handler := handlers.NewUserHandler(logger, Storages.UserStorage)
	handler.Register(router)

	startServer(router, &cfg)
}


func startServer(router *httprouter.Router, cfg *config.Config) {
	logger := logging.GetLogger()
	logger.Info("Start application")

	var listener net.Listener
	var listenErr error
	
	logger.Info("Listen TCP")
	listener, listenErr = net.Listen("tcp",fmt.Sprintf("%s:%s",cfg.Listen.BindIP, cfg.Listen.Port))
	logger.Infof("Server is listening port %s:%s", cfg.Listen.BindIP, cfg.Listen.Port)
	
	if listenErr != nil {
		logger.Fatal(listenErr)
	}
	
	server := &http.Server{
		Handler: router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout: 15 * time.Second,
	}

	logger.Fatal(server.Serve(listener))
}