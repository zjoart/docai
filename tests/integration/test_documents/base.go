package test_documents

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/zjoart/docai/internal/config"
	"github.com/zjoart/docai/internal/database"
	"github.com/zjoart/docai/internal/documents"
	"github.com/zjoart/docai/internal/documents/analyzer"
	"github.com/zjoart/docai/internal/storage"
	"gorm.io/gorm"
)

type TestEnv struct {
	DB      *gorm.DB
	Router  *mux.Router
	Storage *storage.Client
}

func SetupTestEnv(t *testing.T) *TestEnv {

	_ = godotenv.Load("../../../.env")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.Connect(cfg.DBURL)
	if err != nil {
		t.Fatalf("DB connect failed: %v", err)
	}

	minioClient, err := storage.NewMinioClient(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket)
	if err != nil {
		t.Fatalf("Minio init failed: %v", err)
	}

	repo := documents.NewRepository(db)
	ai := analyzer.NewAnalyzer(cfg.OpenRouterAPIKey)
	svc := documents.NewService(repo, minioClient, ai)
	h := documents.NewHandler(svc)

	r := mux.NewRouter()
	documents.RegisterRoutes(r, h)

	return &TestEnv{
		DB:      db,
		Router:  r,
		Storage: minioClient,
	}
}
