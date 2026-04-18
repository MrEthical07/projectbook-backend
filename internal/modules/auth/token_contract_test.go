package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func buildUnsignedJWT(exp int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"exp":%d}`, exp)))
	return header + "." + payload + "."
}

func TestBuildAuthTokenResponseCanonicalContract(t *testing.T) {
	t.Parallel()

	accessToken := buildUnsignedJWT(1_900_000_000)
	result, err := buildAuthTokenResponse(accessToken, "refresh_token_123")
	if err != nil {
		t.Fatalf("buildAuthTokenResponse: %v", err)
	}

	if result.AccessToken != accessToken {
		t.Fatalf("access_token mismatch")
	}
	if result.RefreshToken != "refresh_token_123" {
		t.Fatalf("refresh_token mismatch")
	}
	if result.AccessExpiresUnix != 1_900_000_000 {
		t.Fatalf("access_expires_unix=%d want=%d", result.AccessExpiresUnix, int64(1_900_000_000))
	}

	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	jsonText := string(encoded)
	for _, expected := range []string{"\"access_token\"", "\"refresh_token\"", "\"access_expires_unix\""} {
		if !strings.Contains(jsonText, expected) {
			t.Fatalf("missing expected key %s in %s", expected, jsonText)
		}
	}
	if strings.Contains(jsonText, "access_expires_utc") {
		t.Fatalf("legacy access_expires_utc key must not be present: %s", jsonText)
	}
}
