// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"reviewService/internal/biz"
	"reviewService/internal/conf"
	"reviewService/internal/data"
	"reviewService/internal/server"
	"reviewService/internal/service"
)

import (
	_ "go.uber.org/automaxprocs"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(confServer *conf.Server, confData *conf.Data, consul *conf.Consul, es *conf.ES, logger log.Logger) (*kratos.App, func(), error) {
	registrar := server.NewRegistrar(consul)
	db, err := data.NewDB(confData)
	if err != nil {
		return nil, nil, err
	}
	typedClient, err := data.NewESClient(es)
	if err != nil {
		return nil, nil, err
	}
	client := data.NewRedisClient(confData)
	dataData, cleanup, err := data.NewData(confData, db, typedClient, client, logger)
	if err != nil {
		return nil, nil, err
	}
	reviewRepo := data.NewReviewRepo(dataData, logger)
	reviewUsecase := biz.NewReviewUsecase(reviewRepo, logger)
	reviewService := service.NewReviewService(reviewUsecase)
	grpcServer := server.NewGRPCServer(confServer, reviewService, logger)
	httpServer := server.NewHTTPServer(confServer, logger)
	app := newApp(logger, registrar, grpcServer, httpServer)
	return app, func() {
		cleanup()
	}, nil
}
