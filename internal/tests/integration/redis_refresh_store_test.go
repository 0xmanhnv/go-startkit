//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	auth "appsechub/internal/infras/auth"
)

func TestRedisRefreshStore_IssueValidateRevoke(t *testing.T) {
	ctx := context.Background()
	addr := getenvOr("REDIS_ADDR", "localhost:6379")
	pass := getenvOr("REDIS_PASSWORD", "")
	db := 0

	store := auth.NewRedisRefreshStore(addr, pass, db)
	tok, err := store.Issue(ctx, "user-1", 60)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	uid, err := store.Validate(ctx, tok)
	if err != nil || uid == "" {
		t.Fatalf("validate: %v uid=%s", err, uid)
	}

	if err := store.Revoke(ctx, tok); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	// small delay to allow Redis ops
	time.Sleep(50 * time.Millisecond)
	if _, err := store.Validate(ctx, tok); err == nil {
		t.Fatalf("expected error after revoke")
	}
}

func getenvOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
