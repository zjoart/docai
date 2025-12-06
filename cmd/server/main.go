package main

import (
	"context"
	"log"
	"net/http"

	"github.com/zjoart/docai/internal/config"
	"github.com/zjoart/docai/internal/database"
	"github.com/zjoart/docai/internal/storage"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg.DBURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	minioClient, err := storage.NewMinioClient(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket)
	if err != nil {
		log.Fatalf("Failed to init Minio: %v", err)
	}

	if err := minioClient.EnsureBucket(context.Background()); err != nil {
		log.Fatalf("Failed to ensure bucket: %v", err)
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
