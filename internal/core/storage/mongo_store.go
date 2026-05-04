package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDocumentStore executes document operations over a Mongo database.
type MongoDocumentStore struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoDocumentStore creates a document store backed by Mongo.
func NewMongoDocumentStore(client *mongo.Client, database *mongo.Database) (*MongoDocumentStore, error) {
	if client == nil {
		return nil, errors.New("nil mongo client")
	}
	if database == nil {
		return nil, errors.New("nil mongo database")
	}
	return &MongoDocumentStore{client: client, database: database}, nil
}

// Kind identifies this store as document-oriented.
func (s *MongoDocumentStore) Kind() Kind {
	return KindDocument
}

// WithTx executes a callback in a Mongo transaction scope.
func (s *MongoDocumentStore) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if s == nil || s.client == nil {
		return errors.New("document store is not configured")
	}
	if fn == nil {
		return errors.New("nil transaction callback")
	}

	session, err := s.client.StartSession()
	if err != nil {
		return fmt.Errorf("start mongo session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessionCtx mongo.SessionContext) (any, error) {
		return nil, fn(sessionCtx)
	})
	if err != nil {
		return fmt.Errorf("execute mongo transaction: %w", err)
	}

	return nil
}

// Execute runs one document operation using the Mongo executor.
func (s *MongoDocumentStore) Execute(ctx context.Context, op DocumentOperation) error {
	if s == nil || s.database == nil {
		return errors.New("document store is not configured")
	}
	if op == nil {
		return errors.New("nil document operation")
	}

	exec := mongoDocumentExecutor{database: s.database}
	return op.ExecuteDocument(ctx, exec)
}

// Close disconnects the backing mongo client.
func (s *MongoDocumentStore) Close(ctx context.Context) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Disconnect(ctx)
}

type mongoDocumentExecutor struct {
	database *mongo.Database
}

func (e mongoDocumentExecutor) Run(ctx context.Context, command string, payload any, out any) error {
	if e.database == nil {
		return errors.New("document executor is not configured")
	}

	collectionName, operation, err := parseMongoDocumentCommand(command)
	if err != nil {
		return err
	}

	collection := e.database.Collection(collectionName)

	switch operation {
	case "insert_one":
		result, err := collection.InsertOne(ctx, payload)
		if err != nil {
			return err
		}
		if outMap, ok := out.(map[string]any); ok {
			outMap["inserted_id"] = result.InsertedID
		}
		return nil

	case "find_one":
		filter := extractMongoFilterPayload(payload)
		result := collection.FindOne(ctx, filter)

		if err := result.Err(); err != nil {
			return err
		}
		if out == nil {
			return nil
		}

		var raw map[string]any
		if err := result.Decode(&raw); err != nil {
			return err
		}

		normalized := normalizeBSON(raw)

		if outMap, ok := out.(*map[string]any); ok {
			*outMap = normalized.(map[string]any)
			return nil
		}

		return fmt.Errorf("unsupported output type for find_one")

	case "find_many":
		findPayload := parseMongoFindPayload(payload)

		cursor, err := collection.Find(ctx, findPayload.filter, findPayload.options...)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		if out == nil {
			return nil
		}

		var raw []map[string]any
		if err := cursor.All(ctx, &raw); err != nil {
			return err
		}

		normalized := make([]map[string]any, len(raw))
		for i, doc := range raw {
			normalized[i] = normalizeBSON(doc).(map[string]any)
		}

		if outSlice, ok := out.(*[]map[string]any); ok {
			*outSlice = normalized
			return nil
		}

		return fmt.Errorf("unsupported output type for find_many")

	case "update_one":
		mutationPayload, err := parseMongoMutationPayload(payload)
		if err != nil {
			return err
		}
		result, err := collection.UpdateOne(ctx, mutationPayload.filter, mutationPayload.update, mutationPayload.options...)
		if err != nil {
			return err
		}
		if outMap, ok := out.(map[string]any); ok {
			outMap["matched_count"] = result.MatchedCount
			outMap["modified_count"] = result.ModifiedCount
			outMap["upserted_count"] = result.UpsertedCount
			outMap["upserted_id"] = result.UpsertedID
		}
		return nil

	case "replace_one":
		replacePayload, err := parseMongoReplacePayload(payload)
		if err != nil {
			return err
		}
		result, err := collection.ReplaceOne(ctx, replacePayload.filter, replacePayload.document, replacePayload.options...)
		if err != nil {
			return err
		}
		if outMap, ok := out.(map[string]any); ok {
			outMap["matched_count"] = result.MatchedCount
			outMap["modified_count"] = result.ModifiedCount
			outMap["upserted_count"] = result.UpsertedCount
			outMap["upserted_id"] = result.UpsertedID
		}
		return nil

	case "delete_one":
		filter := extractMongoFilterPayload(payload)
		result, err := collection.DeleteOne(ctx, filter)
		if err != nil {
			return err
		}
		if outMap, ok := out.(map[string]any); ok {
			outMap["deleted_count"] = result.DeletedCount
		}
		return nil

	default:
		return fmt.Errorf("unsupported mongo document operation %q", operation)
	}
}

func parseMongoDocumentCommand(command string) (string, string, error) {
	trimmedCommand := strings.TrimSpace(command)
	if trimmedCommand == "" {
		return "", "", errors.New("empty document command")
	}

	separator := ":"
	if !strings.Contains(trimmedCommand, separator) {
		separator = "."
	}

	parts := strings.SplitN(trimmedCommand, separator, 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid document command %q, expected <collection>:<operation>", trimmedCommand)
	}

	collectionName := strings.TrimSpace(parts[0])
	operation := strings.TrimSpace(parts[1])
	if collectionName == "" || operation == "" {
		return "", "", fmt.Errorf("invalid document command %q, expected non-empty collection and operation", trimmedCommand)
	}

	return collectionName, operation, nil
}

func extractMongoFilterPayload(payload any) any {
	if payload == nil {
		return bson.D{}
	}

	switch val := payload.(type) {
	case map[string]any:
		if filter, exists := val["filter"]; exists {
			return filter
		}
		return val
	case bson.M:
		if filter, exists := val["filter"]; exists {
			return filter
		}
		return val
	default:
		return payload
	}
}

type mongoMutationPayload struct {
	filter  any
	update  any
	options []*options.UpdateOptions
}

func parseMongoMutationPayload(payload any) (mongoMutationPayload, error) {
	var payloadMap map[string]any

	switch val := payload.(type) {
	case map[string]any:
		payloadMap = val
	case bson.M:
		payloadMap = map[string]any(val)
	default:
		return mongoMutationPayload{}, errors.New("update_one payload must be map-like")
	}

	filter, hasFilter := payloadMap["filter"]
	update, hasUpdate := payloadMap["update"]
	if !hasFilter || !hasUpdate {
		return mongoMutationPayload{}, errors.New("update_one payload requires filter and update")
	}

	mutationPayload := mongoMutationPayload{filter: filter, update: update}
	if opt, ok := payloadMap["options"].(*options.UpdateOptions); ok && opt != nil {
		mutationPayload.options = []*options.UpdateOptions{opt}
	}

	return mutationPayload, nil
}

type mongoReplacePayload struct {
	filter   any
	document any
	options  []*options.ReplaceOptions
}

func parseMongoReplacePayload(payload any) (mongoReplacePayload, error) {
	var payloadMap map[string]any

	switch val := payload.(type) {
	case map[string]any:
		payloadMap = val
	case bson.M:
		payloadMap = map[string]any(val)
	default:
		return mongoReplacePayload{}, errors.New("replace_one payload must be map-like")
	}

	filter, hasFilter := payloadMap["filter"]
	document, hasDocument := payloadMap["document"]
	if !hasFilter || !hasDocument {
		return mongoReplacePayload{}, errors.New("replace_one payload requires filter and document")
	}

	replacePayload := mongoReplacePayload{filter: filter, document: document}
	if opt, ok := payloadMap["options"].(*options.ReplaceOptions); ok && opt != nil {
		replacePayload.options = []*options.ReplaceOptions{opt}
	}

	return replacePayload, nil
}

type mongoFindPayload struct {
	filter  any
	options []*options.FindOptions
}

func parseMongoFindPayload(payload any) mongoFindPayload {
	if payload == nil {
		return mongoFindPayload{filter: bson.D{}}
	}

	var payloadMap map[string]any

	switch val := payload.(type) {
	case map[string]any:
		payloadMap = val
	case bson.M:
		payloadMap = map[string]any(val)
	default:
		return mongoFindPayload{filter: payload}
	}

	findPayload := mongoFindPayload{filter: bson.D{}}
	if filter, exists := payloadMap["filter"]; exists {
		findPayload.filter = filter
	}
	if opt, ok := payloadMap["options"].(*options.FindOptions); ok && opt != nil {
		findPayload.options = []*options.FindOptions{opt}
	}

	return findPayload
}

func normalizeBSON(v any) any {
	switch val := v.(type) {
	case bson.M:
		out := make(map[string]any)
		for k, v := range val {
			out[k] = normalizeBSON(v)
		}
		return out

	case map[string]any:
		out := make(map[string]any)
		for k, v := range val {
			out[k] = normalizeBSON(v)
		}
		return out

	case primitive.A:
		out := make([]any, len(val))
		for i, v := range val {
			out[i] = normalizeBSON(v)
		}
		return out

	case []any:
		out := make([]any, len(val))
		for i, v := range val {
			out[i] = normalizeBSON(v)
		}
		return out

	default:
		return val
	}
}
