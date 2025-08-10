package userusecase

import (
	"context"
	"testing"
	"time"

	"gostartkit/internal/application/dto"
	"gostartkit/internal/application/ports"
	domuser "gostartkit/internal/domain/user"

	"github.com/google/uuid"
)

type fakeRepo struct{ user *domuser.User }

func (f *fakeRepo) Save(ctx context.Context, u *domuser.User) error { f.user = u; return nil }
func (f *fakeRepo) GetByID(ctx context.Context, id uuid.UUID) (*domuser.User, error) {
	return f.user, nil
}
func (f *fakeRepo) GetByEmail(ctx context.Context, email domuser.Email) (*domuser.User, error) {
	return f.user, nil
}
func (f *fakeRepo) GetAll(ctx context.Context) ([]*domuser.User, error) {
	return []*domuser.User{f.user}, nil
}
func (f *fakeRepo) Update(ctx context.Context, u *domuser.User) error { f.user = u; return nil }
func (f *fakeRepo) Delete(ctx context.Context, id uuid.UUID) error    { return nil }

type fakeHasher struct{}

func (fakeHasher) Hash(raw string) string                 { return "hashed:" + raw }
func (fakeHasher) Compare(hashed string, raw string) bool { return hashed == "hashed:"+raw }

type fakeTokenIssuer struct{}

func (fakeTokenIssuer) GenerateToken(userID string, role string) (string, error) {
	return "token:" + userID, nil
}

type fakeRefreshStore struct{ issued map[string]time.Time }

func (s *fakeRefreshStore) Issue(ctx context.Context, userID string, ttlSeconds int) (string, error) {
	if s.issued == nil {
		s.issued = make(map[string]time.Time)
	}
	tok := "rftoken:" + userID
	s.issued[tok] = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	return tok, nil
}
func (s *fakeRefreshStore) Rotate(ctx context.Context, oldToken string, ttlSeconds int) (string, string, error) {
	return "", "", nil
}
func (s *fakeRefreshStore) Revoke(ctx context.Context, token string) error {
	delete(s.issued, token)
	return nil
}
func (s *fakeRefreshStore) Validate(ctx context.Context, token string) (string, error) {
	return "", nil
}

var _ ports.TokenIssuer = (*fakeTokenIssuer)(nil)

func TestLoginUserUseCase_Success(t *testing.T) {
	// Arrange
	uid := uuid.New()
	u := &domuser.User{ID: uid, FirstName: "John", LastName: "Doe", Email: domuser.Email("john@example.com"), Password: "hashed:pass", Role: domuser.Role("user"), CreatedAt: time.Now()}
	repo := &fakeRepo{user: u}
	hasher := fakeHasher{}
	jwt := fakeTokenIssuer{}
	store := &fakeRefreshStore{}
	uc := &LoginUserUseCase{repo: repo, hasher: hasher, jwt: jwt, store: store, refreshTTLSeconds: 60}

	// Act
	resp, err := uc.Execute(context.Background(), dto.LoginRequest{Email: "john@example.com", Password: "pass"})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Fatalf("expected access token")
	}
	if resp.RefreshToken == "" {
		t.Fatalf("expected refresh token")
	}
	if resp.User.Email != "john@example.com" {
		t.Fatalf("unexpected user email: %s", resp.User.Email)
	}
}

func TestLoginUserUseCase_InvalidPassword(t *testing.T) {
	uid := uuid.New()
	u := &domuser.User{ID: uid, FirstName: "John", LastName: "Doe", Email: domuser.Email("john@example.com"), Password: "hashed:pass", Role: domuser.Role("user"), CreatedAt: time.Now()}
	repo := &fakeRepo{user: u}
	hasher := fakeHasher{}
	jwt := fakeTokenIssuer{}
	uc := &LoginUserUseCase{repo: repo, hasher: hasher, jwt: jwt}

	_, err := uc.Execute(context.Background(), dto.LoginRequest{Email: "john@example.com", Password: "wrong"})
	if err == nil {
		t.Fatalf("expected error for invalid credentials")
	}
}
