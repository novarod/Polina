package auth_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	appauth "github.com/novarod/polina/apps/api/internal/application/auth"
	"github.com/novarod/polina/apps/api/internal/application/token"
	"github.com/novarod/polina/apps/api/internal/domain/member"
	"github.com/novarod/polina/apps/api/internal/ports"
	"github.com/novarod/polina/apps/api/pkg/apierr"
)

// --- fakes (implement the ports interfaces, no mock lib needed) ---

var (
	_ ports.UserRepository   = (*fakeUserRepo)(nil)
	_ ports.MemberRepository = (*fakeMemberRepo)(nil)
)

type fakeUserRepo struct {
	users     map[string]ports.User // keyed by stored email
	createErr error
	created   *ports.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[string]ports.User)}
}

func (f *fakeUserRepo) Create(_ context.Context, u ports.User) (ports.User, error) {
	if f.createErr != nil {
		return ports.User{}, f.createErr
	}
	f.users[u.Email] = u
	cp := u
	f.created = &cp
	return u, nil
}

func (f *fakeUserRepo) FindByEmail(_ context.Context, email string) (ports.User, error) {
	if u, ok := f.users[email]; ok {
		return u, nil
	}
	return ports.User{}, apierr.NotFound("user")
}

func (f *fakeUserRepo) FindByID(_ context.Context, id uuid.UUID) (ports.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return ports.User{}, apierr.NotFound("user")
}

type fakeMemberRepo struct {
	member  ports.Member
	findErr error
}

func (f *fakeMemberRepo) Create(_ context.Context, m ports.Member) (ports.Member, error) {
	return m, nil
}

func (f *fakeMemberRepo) FindByUserAndOrg(_ context.Context, _, _ uuid.UUID) (ports.Member, error) {
	if f.findErr != nil {
		return ports.Member{}, f.findErr
	}
	return f.member, nil
}

func (f *fakeMemberRepo) SoftDeleteByOrg(_ context.Context, _ uuid.UUID) error { return nil }

// --- helpers ---

func storedUser(t *testing.T, email, password string) ports.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	return ports.User{ID: uuid.New(), Email: email, Name: "Alice", Password: string(hash)}
}

func parseClaims(t *testing.T, tokenStr, secret string) *token.Claims {
	t.Helper()
	claims := &token.Claims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(*jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)
	require.True(t, tok.Valid)
	return claims
}

func appErrCode(t *testing.T, err error) int {
	t.Helper()
	var appErr *apierr.AppError
	require.True(t, errors.As(err, &appErr), "expected *apierr.AppError, got %T", err)
	return appErr.Code
}

// --- Register ---

func TestRegister_DuplicateEmail(t *testing.T) {
	users := newFakeUserRepo()
	users.users["a@b.com"] = ports.User{ID: uuid.New(), Email: "a@b.com"}
	uc := appauth.NewRegisterUseCase(users, bcrypt.MinCost)

	_, err := uc.Execute(context.Background(), appauth.RegisterInput{
		Name: "Al", Email: "a@b.com", Password: "password123",
	})

	require.Error(t, err)
	assert.Equal(t, http.StatusUnprocessableEntity, appErrCode(t, err))
}

func TestRegister_Success_NormalizesAndHashes(t *testing.T) {
	users := newFakeUserRepo()
	uc := appauth.NewRegisterUseCase(users, bcrypt.MinCost)

	out, err := uc.Execute(context.Background(), appauth.RegisterInput{
		Name: "  Alice  ", Email: "Alice@B.com", Password: "password123",
	})

	require.NoError(t, err)
	assert.Equal(t, "alice@b.com", out.Email, "email should be lowercased and trimmed")
	assert.Equal(t, "Alice", out.Name, "name should be trimmed")

	require.NotNil(t, users.created)
	assert.NotEqual(t, "password123", users.created.Password, "password must be hashed, not stored in plaintext")
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(users.created.Password), []byte("password123")))
}

// --- Login ---

func TestLogin_WrongPassword(t *testing.T) {
	users := newFakeUserRepo()
	users.users["a@b.com"] = storedUser(t, "a@b.com", "correct-horse")
	uc := appauth.NewLoginUseCase(users, &fakeMemberRepo{}, "secret", 24)

	_, err := uc.Execute(context.Background(), appauth.LoginInput{Email: "a@b.com", Password: "wrong"})

	require.ErrorIs(t, err, apierr.ErrBadLogin)
}

func TestLogin_UnknownEmail(t *testing.T) {
	uc := appauth.NewLoginUseCase(newFakeUserRepo(), &fakeMemberRepo{}, "secret", 24)

	_, err := uc.Execute(context.Background(), appauth.LoginInput{Email: "ghost@b.com", Password: "whatever"})

	require.ErrorIs(t, err, apierr.ErrBadLogin)
}

func TestLogin_Success_NoOrg(t *testing.T) {
	users := newFakeUserRepo()
	u := storedUser(t, "a@b.com", "correct-horse")
	users.users["a@b.com"] = u
	uc := appauth.NewLoginUseCase(users, &fakeMemberRepo{}, "secret", 24)

	out, err := uc.Execute(context.Background(), appauth.LoginInput{Email: "a@b.com", Password: "correct-horse"})

	require.NoError(t, err)
	assert.Equal(t, u.ID, out.UserID)
	require.NotEmpty(t, out.Token)

	claims := parseClaims(t, out.Token, "secret")
	assert.Equal(t, u.ID, claims.UserID)
	assert.Equal(t, member.Role(""), claims.Role, "role should be empty when no org is selected")
}

func TestLogin_Success_WithOrgLoadsMembership(t *testing.T) {
	users := newFakeUserRepo()
	u := storedUser(t, "a@b.com", "correct-horse")
	users.users["a@b.com"] = u

	orgID := uuid.New()
	memberID := uuid.New()
	members := &fakeMemberRepo{member: ports.Member{
		ID: memberID, UserID: u.ID, OrganizationID: orgID, Role: member.RoleAdmin,
	}}
	uc := appauth.NewLoginUseCase(users, members, "secret", 24)

	out, err := uc.Execute(context.Background(), appauth.LoginInput{
		Email: "a@b.com", Password: "correct-horse", OrganizationID: orgID.String(),
	})

	require.NoError(t, err)
	claims := parseClaims(t, out.Token, "secret")
	assert.Equal(t, orgID, claims.OrgID)
	assert.Equal(t, memberID, claims.MemberID)
	assert.Equal(t, member.RoleAdmin, claims.Role)
}

func TestLogin_NotAMemberOfOrg(t *testing.T) {
	users := newFakeUserRepo()
	u := storedUser(t, "a@b.com", "correct-horse")
	users.users["a@b.com"] = u
	members := &fakeMemberRepo{findErr: apierr.NotFound("member")}
	uc := appauth.NewLoginUseCase(users, members, "secret", 24)

	_, err := uc.Execute(context.Background(), appauth.LoginInput{
		Email: "a@b.com", Password: "correct-horse", OrganizationID: uuid.New().String(),
	})

	require.Error(t, err)
	assert.Equal(t, http.StatusForbidden, appErrCode(t, err))
}
