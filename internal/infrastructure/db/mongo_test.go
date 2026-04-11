package db

import (
	"context"
	"testing"

	"github.com/MrEthical07/superapi/internal/core/config"
)

func TestNewMongoClientRejectsDisabledConfig(t *testing.T) {
	_, err := NewMongoClient(context.Background(), config.MongoConfig{Enabled: false})
	if err == nil {
		t.Fatalf("expected error when mongo is disabled")
	}
}

func TestNewMongoDatabaseRejectsNilClient(t *testing.T) {
	_, err := NewMongoDatabase(nil, "projectbook")
	if err == nil {
		t.Fatalf("expected error for nil mongo client")
	}
}

func TestBootstrapMongoProjectBookCollectionsRejectsNilDatabase(t *testing.T) {
	err := BootstrapMongoProjectBookCollections(context.Background(), nil)
	if err == nil {
		t.Fatalf("expected error for nil mongo database")
	}
}
