package main

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"

	docs "PrioQueue/docs"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func run(
	ctx context.Context,
	logger *slog.Logger,
	args []string,
	getenv func(string) string,
	stdin io.Reader,
	stdout, strerr io.Writer,
) error {
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	docs.SwaggerInfo.Title = "http Priority Queue"
	docs.SwaggerInfo.Description = "This api has endpoints for managing a priority queue"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"

	db, err := initDb()
	if err != nil {
		// This would be a big problem
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close() // Some of the db stuff could get moved elsewhere, but I don't know how to correctly defer closing it if it's in another function?

	v1 := r.Group("/api/v1")
	addRoutes(v1, db, logger)

	err = r.Run("localhost:8080")
	return err
}

func setupLogging() *slog.Logger {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return logger
}

func main() {
	// Setup
	logger := setupLogging()
	ctx := context.Background()

	// Run http server
	err := run(ctx, logger, nil, nil, nil, nil, nil)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(2)
	}
}
