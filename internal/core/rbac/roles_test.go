package rbac

import "testing"

func TestDefaultRoleMask(t *testing.T) {
	tests := []struct {
		role string
		want uint64
		ok   bool
	}{
		{role: RoleOwner, want: 1152921504606846975, ok: true},
		{role: RoleAdmin, want: 864691128455135229, ok: true},
		{role: RoleEditor, want: 20016033248999873, ok: true},
		{role: RoleMember, want: 875734824153537, ok: true},
		{role: RoleViewer, want: 18300341342965825, ok: true},
		{role: RoleLimitedAccess, want: 0, ok: true},
		{role: "unknown", want: 0, ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.role, func(t *testing.T) {
			got, ok := DefaultRoleMask(tc.role)
			if ok != tc.ok {
				t.Fatalf("ok=%t want=%t", ok, tc.ok)
			}
			if got != tc.want {
				t.Fatalf("mask=%d want=%d", got, tc.want)
			}
		})
	}
}

func TestCanonicalRolesReturnsCopy(t *testing.T) {
	roles := CanonicalRoles()
	if len(roles) != 6 {
		t.Fatalf("len(roles)=%d want=6", len(roles))
	}

	roles[0] = "mutated"
	again := CanonicalRoles()
	if again[0] != RoleOwner {
		t.Fatalf("expected canonical roles to be immutable copy")
	}
}
