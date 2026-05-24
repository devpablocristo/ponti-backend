package admin

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// --- fake IDPClient para no depender del NoopAdmin (que retorna error) ---

type fakeIDP struct {
	uidByEmail map[string]string
	createErr  error
}

func (f *fakeIDP) CreateUser(_ context.Context, email, _ string) (string, error) {
	if f.createErr != nil {
		return "", f.createErr
	}
	if uid, ok := f.uidByEmail[email]; ok {
		return uid, nil
	}
	uid := "uid-" + email
	if f.uidByEmail == nil {
		f.uidByEmail = make(map[string]string)
	}
	f.uidByEmail[email] = uid
	return uid, nil
}

func (f *fakeIDP) GetUserUIDByEmail(_ context.Context, email string) (string, error) {
	if uid, ok := f.uidByEmail[email]; ok {
		return uid, nil
	}
	return "", errors.New("not found")
}

func (f *fakeIDP) GeneratePasswordResetLink(_ context.Context, email string) (string, error) {
	return "https://reset/" + email, nil
}

// --- fake RepositoryPort para aislar UseCases sin gormtest ---

type fakeRepo struct {
	tenants     []Tenant
	userByIDP   map[string]*LocalUser
	tenantByID  map[string]uuid.UUID
	roleByName  map[string]uuid.UUID
	memberships []MeMembership
	perms       []RolePermission

	createInviteCalls int
	upsertCalls       int
}

func (r *fakeRepo) EnsureLocalUserByIDPSub(_ context.Context, idpSub, email string) (*LocalUser, error) {
	if u, ok := r.userByIDP[idpSub]; ok {
		return u, nil
	}
	u := &LocalUser{ID: uuid.New(), Email: email, IDPSub: idpSub}
	if r.userByIDP == nil {
		r.userByIDP = make(map[string]*LocalUser)
	}
	r.userByIDP[idpSub] = u
	return u, nil
}

func (r *fakeRepo) GetLocalUserByIDPSub(_ context.Context, idpSub string) (*LocalUser, error) {
	if u, ok := r.userByIDP[idpSub]; ok {
		return u, nil
	}
	return nil, errors.New("not found")
}

func (r *fakeRepo) EnsureTenantByName(_ context.Context, name string) (uuid.UUID, error) {
	if id, ok := r.tenantByID[name]; ok {
		return id, nil
	}
	id := uuid.New()
	if r.tenantByID == nil {
		r.tenantByID = make(map[string]uuid.UUID)
	}
	r.tenantByID[name] = id
	r.tenants = append(r.tenants, Tenant{ID: id, Name: name})
	return id, nil
}

func (r *fakeRepo) RoleIDByName(_ context.Context, name string) (uuid.UUID, error) {
	if id, ok := r.roleByName[name]; ok {
		return id, nil
	}
	return uuid.Nil, errors.New("role not found")
}

func (r *fakeRepo) UpsertMembership(_ context.Context, _, _, _ uuid.UUID) error {
	r.upsertCalls++
	return nil
}

func (r *fakeRepo) ListTenants(_ context.Context) ([]Tenant, error) {
	return r.tenants, nil
}

func (r *fakeRepo) ListUsersForTenant(_ context.Context, _ uuid.UUID) ([]UserMembership, error) {
	return nil, nil
}

func (r *fakeRepo) CreateInvite(_ context.Context, tenantID uuid.UUID, email string, roleID uuid.UUID, tokenHash string, expiresAt time.Time, invitedBy uuid.UUID) (*TenantInvite, error) {
	r.createInviteCalls++
	return &TenantInvite{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Email:     email,
		RoleID:    roleID,
		ExpiresAt: expiresAt,
		InvitedBy: &invitedBy,
	}, nil
}

func (r *fakeRepo) AcceptInvite(_ context.Context, _ string, _ uuid.UUID) (*TenantInvite, error) {
	return &TenantInvite{ID: uuid.New(), TenantID: uuid.New(), RoleID: uuid.New()}, nil
}

func (r *fakeRepo) UpdateMembershipRole(_ context.Context, _, _, _ uuid.UUID) error {
	return nil
}

func (r *fakeRepo) ArchiveMembership(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (r *fakeRepo) ListMembershipsForUser(_ context.Context, _ uuid.UUID) ([]MeMembership, error) {
	return r.memberships, nil
}

func (r *fakeRepo) ListPermissionsByRoleIDs(_ context.Context, _ []uuid.UUID) ([]RolePermission, error) {
	return r.perms, nil
}

// --- tests ---

func TestUsernameToEmail(t *testing.T) {
	cases := map[string]string{
		"":              "",
		"  ":            "",
		"pablo":         "pablo@ponti.local",
		"pablo@foo.com": "pablo@foo.com",
		"  pablo  ":     "pablo@ponti.local",
	}
	for input, expected := range cases {
		if got := UsernameToEmail(input); got != expected {
			t.Errorf("UsernameToEmail(%q): got %q, want %q", input, got, expected)
		}
	}
}

func TestHashInviteToken_StableAndCaseSensitiveTrim(t *testing.T) {
	a := hashInviteToken("token-abc")
	b := hashInviteToken("  token-abc  ")
	if a != b {
		t.Fatal("hash should be stable across whitespace trim")
	}
	c := hashInviteToken("token-ABC")
	if a == c {
		t.Fatal("hash should be case-sensitive")
	}
	if len(a) != 64 {
		t.Fatalf("sha256 hex should be 64 chars, got %d", len(a))
	}
}

func TestNewInviteToken_NonEmptyHex(t *testing.T) {
	tok, err := newInviteToken()
	if err != nil {
		t.Fatal(err)
	}
	if len(tok) != 64 {
		t.Fatalf("expected 64 hex chars (32 bytes), got %d", len(tok))
	}
	if !isHex(tok) {
		t.Fatalf("token is not hex: %s", tok)
	}
}

func isHex(s string) bool {
	const hex = "0123456789abcdef"
	for _, c := range s {
		if !strings.ContainsRune(hex, c) {
			return false
		}
	}
	return true
}

func TestUseCases_CreateUser_HappyPath(t *testing.T) {
	repo := &fakeRepo{
		roleByName: map[string]uuid.UUID{
			"viewer": uuid.New(),
		},
	}
	uc := NewUseCases(repo, &fakeIDP{})
	out, err := uc.CreateUser(context.Background(), CreateUserInput{
		Email:      "alice@example.com",
		Password:   "s3cret",
		TenantName: "default",
		RoleName:   "viewer",
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if out.User == nil {
		t.Fatal("expected user in output")
	}
	if out.RoleName != "viewer" {
		t.Errorf("expected role viewer, got %q", out.RoleName)
	}
	if repo.upsertCalls != 1 {
		t.Errorf("expected 1 membership upsert, got %d", repo.upsertCalls)
	}
}

func TestUseCases_CreateUser_RejectsEmpty(t *testing.T) {
	uc := NewUseCases(&fakeRepo{}, &fakeIDP{})
	if _, err := uc.CreateUser(context.Background(), CreateUserInput{Email: "", Password: ""}); err == nil {
		t.Fatal("expected validation error when email/password missing")
	}
}

func TestUseCases_CreateInvite_DefaultsRoleAndExpiry(t *testing.T) {
	repo := &fakeRepo{
		roleByName: map[string]uuid.UUID{
			"tenant_viewer": uuid.New(),
		},
	}
	uc := NewUseCases(repo, &fakeIDP{})
	tenantID := uuid.New()
	out, err := uc.CreateInvite(context.Background(), CreateInviteInput{
		TenantID: tenantID,
		Email:    "guest@example.com",
		// RoleName vacío -> debe defaultear a tenant_viewer
		// ExpiresIn vacío -> debe defaultear a 7 días
	})
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if out.Token == "" || len(out.Token) != 64 {
		t.Fatalf("expected 64-char token, got %d", len(out.Token))
	}
	if out.Invite == nil || out.Invite.TenantID != tenantID {
		t.Fatal("expected invite tied to tenant")
	}
	if repo.createInviteCalls != 1 {
		t.Errorf("expected 1 invite create, got %d", repo.createInviteCalls)
	}
}

func TestUseCases_GetMeContext_RequiresActor(t *testing.T) {
	uc := NewUseCases(&fakeRepo{}, &fakeIDP{})
	if _, err := uc.GetMeContext(context.Background(), "", uuid.Nil); err == nil {
		t.Fatal("expected Unauthorized when actor empty")
	}
}

func TestUseCases_GetMeContext_BuildsTenantList(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	roleID := uuid.New()
	repo := &fakeRepo{
		userByIDP: map[string]*LocalUser{
			"sub-1": {ID: userID, Email: "alice@example.com", IDPSub: "sub-1"},
		},
		memberships: []MeMembership{
			{TenantID: tenantID, Name: "default", RoleID: roleID, RoleName: "admin"},
		},
		perms: []RolePermission{
			{RoleID: roleID, Name: "customers.read"},
			{RoleID: roleID, Name: "customers.write"},
		},
	}
	uc := NewUseCases(repo, &fakeIDP{})

	out, err := uc.GetMeContext(context.Background(), "sub-1", tenantID)
	if err != nil {
		t.Fatalf("GetMeContext: %v", err)
	}
	if out.User.ID != userID {
		t.Errorf("user id mismatch: %v vs %v", out.User.ID, userID)
	}
	if out.CurrentTenantID != tenantID {
		t.Error("current tenant id mismatch")
	}
	if len(out.Tenants) != 1 {
		t.Fatalf("expected 1 tenant, got %d", len(out.Tenants))
	}
	tn := out.Tenants[0]
	if !tn.IsCurrent {
		t.Error("expected is_current=true for matching tenant")
	}
	if len(tn.Permissions) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(tn.Permissions))
	}
}
