package domain

import (
	"errors"
	"testing"
	"time"
)

func TestIsValidKind(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"natural_person", "natural_person", true},
		{"organization", "organization", true},
		{"other", "other", true},
		{"unknown", "unknown", true},
		{"with leading space", "  organization", true},
		{"with trailing space", "natural_person  ", true},
		{"empty", "", false},
		{"random string", "rare", false},
		{"upper case rejected", "ORGANIZATION", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsValidKind(tc.in); got != tc.want {
				t.Fatalf("IsValidKind(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestIsValidRole(t *testing.T) {
	for _, role := range []string{"cliente", "responsable", "inversor", "arrendatario", "proveedor", "contratista", "facturador"} {
		if !IsValidRole(role) {
			t.Errorf("IsValidRole(%q) should be true", role)
		}
	}

	for _, role := range []string{"", "admin", "owner", "Cliente", "  "} {
		if IsValidRole(role) {
			t.Errorf("IsValidRole(%q) should be false", role)
		}
	}
}

func TestActorValidate(t *testing.T) {
	t.Run("nil actor returns nil", func(t *testing.T) {
		var a *Actor
		if err := a.Validate(); err != nil {
			t.Fatalf("expected nil for nil actor, got %v", err)
		}
	})

	t.Run("valid kind + valid roles", func(t *testing.T) {
		a := &Actor{
			ActorKind: "organization",
			Roles:     []string{"cliente", "proveedor"},
		}
		if err := a.Validate(); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("invalid kind", func(t *testing.T) {
		a := &Actor{ActorKind: "alien", Roles: []string{"cliente"}}
		err := a.Validate()
		if !errors.Is(err, ErrInvalidActorKind) {
			t.Fatalf("expected ErrInvalidActorKind, got %v", err)
		}
	})

	t.Run("valid kind but invalid role", func(t *testing.T) {
		a := &Actor{ActorKind: "natural_person", Roles: []string{"cliente", "owner"}}
		err := a.Validate()
		if !errors.Is(err, ErrInvalidActorRole) {
			t.Fatalf("expected ErrInvalidActorRole, got %v", err)
		}
	})

	t.Run("empty kind rejected", func(t *testing.T) {
		a := &Actor{ActorKind: "", Roles: nil}
		err := a.Validate()
		if !errors.Is(err, ErrInvalidActorKind) {
			t.Fatalf("expected ErrInvalidActorKind, got %v", err)
		}
	})

	t.Run("empty roles slice is OK", func(t *testing.T) {
		a := &Actor{ActorKind: "organization", Roles: []string{}}
		if err := a.Validate(); err != nil {
			t.Fatalf("expected nil for actor with empty roles, got %v", err)
		}
	})
}

func TestActorIsArchived(t *testing.T) {
	t.Run("nil actor is archived", func(t *testing.T) {
		var a *Actor
		if !a.IsArchived() {
			t.Fatal("nil actor should be archived (defensive)")
		}
	})

	t.Run("no archived_at means active", func(t *testing.T) {
		a := &Actor{ActorKind: "organization"}
		if a.IsArchived() {
			t.Fatal("actor without ArchivedAt should not be archived")
		}
	})

	t.Run("with archived_at means archived", func(t *testing.T) {
		now := time.Now()
		a := &Actor{ActorKind: "organization", ArchivedAt: &now}
		if !a.IsArchived() {
			t.Fatal("actor with ArchivedAt should be archived")
		}
	})
}
