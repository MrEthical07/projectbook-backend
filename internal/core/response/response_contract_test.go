package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apperr "github.com/MrEthical07/superapi/internal/core/errors"
)

func decodeEnvelope(t *testing.T, body string) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	return payload
}

func TestOKEnvelopeContractUsesSnakeCaseRequestID(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	OK(rr, map[string]any{"project_id": "proj_123"}, "req_abc")

	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusOK)
	}

	payload := decodeEnvelope(t, rr.Body.String())
	if _, hasLegacy := payload["requestId"]; hasLegacy {
		t.Fatalf("legacy requestId key should not be present: %v", payload)
	}
	if payload["request_id"] != "req_abc" {
		t.Fatalf("request_id=%v want=req_abc", payload["request_id"])
	}
	if payload["success"] != true {
		t.Fatalf("success=%v want=true", payload["success"])
	}
}

func TestErrorEnvelopeContractIncludesCodeMessageAndRequestID(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	Error(rr, apperr.New(apperr.CodeBadRequest, http.StatusBadRequest, "invalid request"), "req_bad")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusBadRequest)
	}

	payload := decodeEnvelope(t, rr.Body.String())
	if payload["success"] != false {
		t.Fatalf("success=%v want=false", payload["success"])
	}
	if payload["request_id"] != "req_bad" {
		t.Fatalf("request_id=%v want=req_bad", payload["request_id"])
	}
	if _, hasLegacy := payload["requestId"]; hasLegacy {
		t.Fatalf("legacy requestId key should not be present: %v", payload)
	}

	errorValue, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("error payload missing or invalid: %v", payload["error"])
	}
	if errorValue["code"] != string(apperr.CodeBadRequest) {
		t.Fatalf("error.code=%v want=%s", errorValue["code"], apperr.CodeBadRequest)
	}
	if errorValue["message"] != "invalid request" {
		t.Fatalf("error.message=%v want=invalid request", errorValue["message"])
	}
}

func TestUnknownErrorsMapToInternalContract(t *testing.T) {
	t.Parallel()

	rr := httptest.NewRecorder()
	Error(rr, errors.New("db exploded"), "req_internal")

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d want=%d", rr.Code, http.StatusInternalServerError)
	}

	payload := decodeEnvelope(t, rr.Body.String())
	errorValue, ok := payload["error"].(map[string]any)
	if !ok {
		t.Fatalf("error payload missing or invalid: %v", payload["error"])
	}
	if errorValue["code"] != string(apperr.CodeInternal) {
		t.Fatalf("error.code=%v want=%s", errorValue["code"], apperr.CodeInternal)
	}
}
