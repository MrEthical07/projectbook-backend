package storage

import (
	"testing"
)

func TestNewMongoDocumentStoreRejectsNilInputs(t *testing.T) {
	if _, err := NewMongoDocumentStore(nil, nil); err == nil {
		t.Fatalf("expected error for nil client and database")
	}
}

func TestParseMongoDocumentCommand(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		wantCollection string
		wantOperation  string
		wantErr        bool
	}{
		{name: "colon separator", command: "story_documents:find_one", wantCollection: "story_documents", wantOperation: "find_one"},
		{name: "dot separator", command: "story_documents.find_one", wantCollection: "story_documents", wantOperation: "find_one"},
		{name: "invalid format", command: "story_documents", wantErr: true},
		{name: "empty command", command: "", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			collectionName, operation, err := parseMongoDocumentCommand(tc.command)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for command %q", tc.command)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseMongoDocumentCommand(%q) error = %v", tc.command, err)
			}
			if collectionName != tc.wantCollection {
				t.Fatalf("collection=%q want=%q", collectionName, tc.wantCollection)
			}
			if operation != tc.wantOperation {
				t.Fatalf("operation=%q want=%q", operation, tc.wantOperation)
			}
		})
	}
}

func TestParseMongoMutationPayloadRequiresFilterAndUpdate(t *testing.T) {
	if _, err := parseMongoMutationPayload(nil); err == nil {
		t.Fatalf("expected error for nil payload")
	}

	if _, err := parseMongoMutationPayload(map[string]any{"filter": map[string]any{"id": "1"}}); err == nil {
		t.Fatalf("expected error when update is missing")
	}

	payload, err := parseMongoMutationPayload(map[string]any{
		"filter": map[string]any{"id": "1"},
		"update": map[string]any{"$set": map[string]any{"title": "updated"}},
	})
	if err != nil {
		t.Fatalf("parseMongoMutationPayload() error = %v", err)
	}
	if payload.filter == nil || payload.update == nil {
		t.Fatalf("expected non-nil filter and update payload")
	}
}
