package userusecase

import (
	"context"
	"testing"
	"time"

	domuser "gostartkit/internal/domain/user"

	"github.com/google/uuid"
)

type repoOne struct{ u *domuser.User }

func (r repoOne) Save(ctx context.Context, u *domuser.User) error                  { return nil }
func (r repoOne) GetByID(ctx context.Context, id uuid.UUID) (*domuser.User, error) { return r.u, nil }
func (r repoOne) GetByEmail(ctx context.Context, email domuser.Email) (*domuser.User, error) {
	return r.u, nil
}
func (r repoOne) GetAll(ctx context.Context) ([]*domuser.User, error) {
	return []*domuser.User{r.u}, nil
}
func (r repoOne) Update(ctx context.Context, u *domuser.User) error { return nil }
func (r repoOne) Delete(ctx context.Context, id uuid.UUID) error    { return nil }

func TestGetMe_Success(t *testing.T) {
	uid := uuid.New()
	u := &domuser.User{ID: uid, FirstName: "A", LastName: "B", Email: domuser.Email("a@b.com"), CreatedAt: time.Now()}
	usecase := NewGetMeUseCase(repoOne{u: u})
	got, err := usecase.Execute(context.Background(), uid.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != uid {
		t.Fatalf("expected %s, got %s", uid, got.ID)
	}
}
