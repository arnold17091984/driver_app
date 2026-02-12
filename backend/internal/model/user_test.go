package model

import "testing"

func TestRoleIsValid(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{RoleAdmin, true},
		{RoleDispatcher, true},
		{RoleViewer, true},
		{RoleDriver, true},
		{Role("superadmin"), false},
		{Role(""), false},
	}

	for _, tc := range tests {
		t.Run(string(tc.role), func(t *testing.T) {
			if got := tc.role.IsValid(); got != tc.want {
				t.Errorf("Role(%q).IsValid() = %v, want %v", tc.role, got, tc.want)
			}
		})
	}
}

func TestRoleCanDispatch(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{RoleAdmin, true},
		{RoleDispatcher, true},
		{RoleViewer, false},
		{RoleDriver, false},
	}

	for _, tc := range tests {
		t.Run(string(tc.role), func(t *testing.T) {
			if got := tc.role.CanDispatch(); got != tc.want {
				t.Errorf("Role(%q).CanDispatch() = %v, want %v", tc.role, got, tc.want)
			}
		})
	}
}

func TestRoleIsAdmin(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{RoleAdmin, true},
		{RoleDispatcher, false},
		{RoleViewer, false},
		{RoleDriver, false},
	}

	for _, tc := range tests {
		t.Run(string(tc.role), func(t *testing.T) {
			if got := tc.role.IsAdmin(); got != tc.want {
				t.Errorf("Role(%q).IsAdmin() = %v, want %v", tc.role, got, tc.want)
			}
		})
	}
}
