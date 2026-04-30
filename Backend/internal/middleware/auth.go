package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	jwtSecret     string
	revokedTokens map[string]time.Time
	mu            sync.RWMutex
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:     secret,
		revokedTokens: make(map[string]time.Time),
	}
}

func (m *AuthMiddleware) GenerateToken(userID, username, role string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.jwtSecret))
}

func (m *AuthMiddleware) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
		}
		return []byte(m.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func (m *AuthMiddleware) RevokeToken(tokenID string, expiresAt time.Time) {
	if tokenID == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.revokedTokens[tokenID] = expiresAt
	m.cleanupExpiredRevokedTokens(time.Now())
}

func (m *AuthMiddleware) IsTokenRevoked(tokenID string) bool {
	if tokenID == "" {
		return false
	}

	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cleanupExpiredRevokedTokens(now)

	expiresAt, ok := m.revokedTokens[tokenID]
	return ok && expiresAt.After(now)
}

func (m *AuthMiddleware) cleanupExpiredRevokedTokens(now time.Time) {
	for tokenID, expiresAt := range m.revokedTokens {
		if expiresAt.Before(now) {
			delete(m.revokedTokens, tokenID)
		}
	}
}

func ExtractBearerToken(authHeader string) (string, bool) {
	if authHeader == "" {
		return "", false
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader || tokenString == "" {
		return "", false
	}
	return tokenString, true
}

func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		tokenString, ok := ExtractBearerToken(authHeader)
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		claims, err := m.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if m.IsTokenRevoked(claims.ID) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func (m *AuthMiddleware) RoleRequired(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Role not found"})
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, r := range allowedRoles {
			if userRole == r {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}
