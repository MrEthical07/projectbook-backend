package auth

import (
	"testing"

	goauth "github.com/MrEthical07/goAuth"
)

func TestMapUserToRecordIncludesAccountVersion(t *testing.T) {
	row := StoredUser{
		ID:             "user-1",
		Email:          "person@example.com",
		PasswordHash:   "hash",
		AccountVersion: 7,
		EmailVerified:  true,
	}

	record := mapUserToRecord(row)
	if record.UserID != row.ID {
		t.Fatalf("UserID=%q want=%q", record.UserID, row.ID)
	}
	if record.Identifier != row.Email {
		t.Fatalf("Identifier=%q want=%q", record.Identifier, row.Email)
	}
	if record.AccountVersion != row.AccountVersion {
		t.Fatalf("AccountVersion=%d want=%d", record.AccountVersion, row.AccountVersion)
	}
	if record.Status != goauth.AccountActive {
		t.Fatalf("Status=%v want=%v", record.Status, goauth.AccountActive)
	}
}

func TestMapUserToRecordPendingVerification(t *testing.T) {
	record := mapUserToRecord(StoredUser{EmailVerified: false})
	if record.Status != goauth.AccountPendingVerification {
		t.Fatalf("Status=%v want=%v", record.Status, goauth.AccountPendingVerification)
	}
}
