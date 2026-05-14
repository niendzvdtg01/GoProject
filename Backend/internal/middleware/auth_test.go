package middleware

import (
	"testing"
	"time"
)

func newTestAuth() *AuthMiddleware {
	return NewAuthMiddleware("test-secret-key")
}

func TestGenerateToken_ReturnsNonEmptyToken(t *testing.T) {
	auth := newTestAuth()
	token, err := auth.GenerateToken("user-1", "alice", "member")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateToken_Success(t *testing.T) {
	auth := newTestAuth()
	token, _ := auth.GenerateToken("user-1", "alice", "member")

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Errorf("expected user_id=user-1, got %s", claims.UserID)
	}
	if claims.Username != "alice" {
		t.Errorf("expected username=alice, got %s", claims.Username)
	}
	if claims.Role != "member" {
		t.Errorf("expected role=member, got %s", claims.Role)
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	auth := newTestAuth()
	otherAuth := NewAuthMiddleware("different-secret")
	token, _ := otherAuth.GenerateToken("user-1", "alice", "member")

	_, err := auth.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestValidateToken_Revoked(t *testing.T) {
	auth := newTestAuth()
	token, _ := auth.GenerateToken("user-1", "alice", "member")

	claims, _ := auth.ValidateToken(token)
	auth.RevokeToken(claims.ID, claims.ExpiresAt.Time)

	_, err := auth.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for revoked token")
	}
	if err != ErrTokenRevoked {
		t.Errorf("expected ErrTokenRevoked, got %v", err)
	}
}

func TestRevokeToken_EmptyIDIsNoop(t *testing.T) {
	auth := newTestAuth()
	// Should not panic
	auth.RevokeToken("", time.Now().Add(time.Hour))
}

func TestIsTokenRevoked_NotRevoked(t *testing.T) {
	auth := newTestAuth()
	if auth.IsTokenRevoked("some-jti") {
		t.Fatal("token should not be revoked")
	}
}

func TestIsTokenRevoked_Revoked(t *testing.T) {
	auth := newTestAuth()
	auth.RevokeToken("jti-abc", time.Now().Add(time.Hour))
	if !auth.IsTokenRevoked("jti-abc") {
		t.Fatal("token should be revoked")
	}
}

func TestIsTokenRevoked_ExpiredRevocation(t *testing.T) {
	auth := newTestAuth()
	auth.RevokeToken("jti-old", time.Now().Add(-time.Second))
	// Expiry is in the past, so IsTokenRevoked should return false
	if auth.IsTokenRevoked("jti-old") {
		t.Fatal("expired revocation should not mark token as revoked")
	}
}

func TestExtractBearerToken_Valid(t *testing.T) {
	token, ok := ExtractBearerToken("Bearer mytoken123")
	if !ok {
		t.Fatal("expected ok=true")
	}
	if token != "mytoken123" {
		t.Errorf("expected mytoken123, got %s", token)
	}
}

func TestExtractBearerToken_Empty(t *testing.T) {
	_, ok := ExtractBearerToken("")
	if ok {
		t.Fatal("expected ok=false for empty header")
	}
}

func TestExtractBearerToken_NoBearer(t *testing.T) {
	_, ok := ExtractBearerToken("mytoken123")
	if ok {
		t.Fatal("expected ok=false for missing Bearer prefix")
	}
}

func TestExtractBearerToken_BearerOnly(t *testing.T) {
	_, ok := ExtractBearerToken("Bearer ")
	if ok {
		t.Fatal("expected ok=false for empty token after Bearer")
	}
}
