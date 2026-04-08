package db

import "testing"

func TestResolvePostgresURLPrefersConfigValue(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://env")
	t.Setenv("POSTGRES_URL", "postgres://legacy")

	if got, want := resolvePostgresURL("postgres://config"), "postgres://config"; got != want {
		t.Fatalf("resolvePostgresURL()=%q want=%q", got, want)
	}
}

func TestResolvePostgresURLFallsBackToDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://database-url")
	t.Setenv("POSTGRES_URL", "postgres://legacy")

	if got, want := resolvePostgresURL(""), "postgres://database-url"; got != want {
		t.Fatalf("resolvePostgresURL()=%q want=%q", got, want)
	}
}

func TestResolvePostgresURLFallsBackToPostgresURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("POSTGRES_URL", "postgres://legacy")

	if got, want := resolvePostgresURL(""), "postgres://legacy"; got != want {
		t.Fatalf("resolvePostgresURL()=%q want=%q", got, want)
	}
}
