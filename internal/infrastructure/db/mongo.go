package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/MrEthical07/superapi/internal/core/config"
)

type mongoIndexSpec struct {
	name   string
	keys   bson.D
	unique bool
}

type mongoCollectionSpec struct {
	name    string
	indexes []mongoIndexSpec
}

var projectBookMongoCollections = []mongoCollectionSpec{
	{
		name: "story_documents",
		indexes: []mongoIndexSpec{
			{name: "ux_story_documents_artifact_id", keys: bson.D{{Key: "artifact_id", Value: 1}}, unique: true},
			{name: "ix_story_documents_project_updated", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "updated_at", Value: -1}}},
			{name: "ix_story_documents_project_revision", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "revision", Value: -1}}},
		},
	},
	{
		name: "journey_documents",
		indexes: []mongoIndexSpec{
			{name: "ux_journey_documents_artifact_id", keys: bson.D{{Key: "artifact_id", Value: 1}}, unique: true},
			{name: "ix_journey_documents_project_updated", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "updated_at", Value: -1}}},
			{name: "ix_journey_documents_project_revision", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "revision", Value: -1}}},
		},
	},
	{
		name: "problem_documents",
		indexes: []mongoIndexSpec{
			{name: "ux_problem_documents_artifact_id", keys: bson.D{{Key: "artifact_id", Value: 1}}, unique: true},
			{name: "ix_problem_documents_project_updated", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "updated_at", Value: -1}}},
			{name: "ix_problem_documents_project_revision", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "revision", Value: -1}}},
		},
	},
	{
		name: "idea_documents",
		indexes: []mongoIndexSpec{
			{name: "ux_idea_documents_artifact_id", keys: bson.D{{Key: "artifact_id", Value: 1}}, unique: true},
			{name: "ix_idea_documents_project_updated", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "updated_at", Value: -1}}},
			{name: "ix_idea_documents_project_revision", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "revision", Value: -1}}},
		},
	},
	{
		name: "task_documents",
		indexes: []mongoIndexSpec{
			{name: "ux_task_documents_artifact_id", keys: bson.D{{Key: "artifact_id", Value: 1}}, unique: true},
			{name: "ix_task_documents_project_updated", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "updated_at", Value: -1}}},
			{name: "ix_task_documents_project_revision", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "revision", Value: -1}}},
		},
	},
	{
		name: "feedback_documents",
		indexes: []mongoIndexSpec{
			{name: "ux_feedback_documents_artifact_id", keys: bson.D{{Key: "artifact_id", Value: 1}}, unique: true},
			{name: "ix_feedback_documents_project_updated", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "updated_at", Value: -1}}},
			{name: "ix_feedback_documents_project_revision", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "revision", Value: -1}}},
		},
	},
	{
		name: "page_documents",
		indexes: []mongoIndexSpec{
			{name: "ux_page_documents_artifact_id", keys: bson.D{{Key: "artifact_id", Value: 1}}, unique: true},
			{name: "ix_page_documents_project_updated", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "updated_at", Value: -1}}},
			{name: "ix_page_documents_project_revision", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "revision", Value: -1}}},
		},
	},
	{
		name: "resource_documents",
		indexes: []mongoIndexSpec{
			{name: "ux_resource_documents_artifact_id", keys: bson.D{{Key: "artifact_id", Value: 1}}, unique: true},
			{name: "ix_resource_documents_project_updated", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "updated_at", Value: -1}}},
			{name: "ix_resource_documents_project_revision", keys: bson.D{{Key: "project_id", Value: 1}, {Key: "revision", Value: -1}}},
		},
	},
	{
		name: "resource_version_documents",
		indexes: []mongoIndexSpec{
			{name: "ux_resource_version_documents_resource_version_id", keys: bson.D{{Key: "resource_version_id", Value: 1}}, unique: true},
			{name: "ix_resource_version_documents_resource_revision", keys: bson.D{{Key: "resource_id", Value: 1}, {Key: "revision", Value: -1}}},
		},
	},
}

// NewMongoClient creates a mongo client and verifies connectivity with a startup ping.
func NewMongoClient(ctx context.Context, cfg config.MongoConfig) (*mongo.Client, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("mongo is disabled")
	}

	uri := strings.TrimSpace(cfg.URL)
	if uri == "" {
		return nil, fmt.Errorf("mongo url cannot be empty")
	}

	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(uint64(cfg.MaxPoolSize)).
		SetMinPoolSize(uint64(cfg.MinPoolSize)).
		SetConnectTimeout(cfg.ConnectTimeout)

	startupCtx, cancel := context.WithTimeout(ctx, cfg.StartupPingTimeout)
	defer cancel()

	client, err := mongo.Connect(startupCtx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	if err := client.Ping(startupCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	return client, nil
}

// NewMongoDatabase resolves the configured database handle from a client.
func NewMongoDatabase(client *mongo.Client, databaseName string) (*mongo.Database, error) {
	if client == nil {
		return nil, fmt.Errorf("mongo client is nil")
	}
	name := strings.TrimSpace(databaseName)
	if name == "" {
		return nil, fmt.Errorf("mongo database name cannot be empty")
	}
	return client.Database(name), nil
}

// CheckMongoHealth performs a bounded ping for readiness checks.
func CheckMongoHealth(ctx context.Context, client *mongo.Client, timeout time.Duration) error {
	if client == nil {
		return fmt.Errorf("mongo client is nil")
	}
	if timeout <= 0 {
		timeout = time.Second
	}

	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return client.Ping(checkCtx, nil)
}

// BootstrapMongoProjectBookCollections ensures required collections and indexes exist.
func BootstrapMongoProjectBookCollections(ctx context.Context, database *mongo.Database) error {
	if database == nil {
		return fmt.Errorf("mongo database is nil")
	}

	for _, collectionSpec := range projectBookMongoCollections {
		collection := database.Collection(collectionSpec.name)
		for _, indexSpec := range collectionSpec.indexes {
			_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
				Keys: indexSpec.keys,
				Options: options.Index().
					SetName(indexSpec.name).
					SetUnique(indexSpec.unique),
			})
			if err != nil {
				return fmt.Errorf("ensure mongo index %s on %s: %w", indexSpec.name, collectionSpec.name, err)
			}
		}
	}

	return nil
}
