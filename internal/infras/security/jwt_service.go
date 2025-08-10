package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
	GenerateToken(userID string, role string) (string, error)
	ValidateToken(tokenStr string) (*AppClaims, error)
}

type jwtService struct {
	secretKey      string
	expireDuration time.Duration
	issuer         string
	audience       string
	leeway         time.Duration
}

func NewJWTService(secret string, expireSec int) JWTService {
	return &jwtService{
		secretKey:      secret,
		expireDuration: time.Duration(expireSec) * time.Second,
	}
}

// SetMeta configures issuer, audience, and leeway for validation.
func (j *jwtService) SetMeta(iss, aud string, leewaySec int) {
	j.issuer = iss
	j.audience = aud
	if leewaySec > 0 {
		j.leeway = time.Duration(leewaySec) * time.Second
	}
}

// AppClaims includes standard registered claims plus application-specific fields
type AppClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

func (j *jwtService) GenerateToken(userID string, role string) (string, error) {
	now := time.Now()
	claims := AppClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(j.expireDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Audience:  jwt.ClaimStrings{j.audience},
		},
		Role: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *jwtService) ValidateToken(tokenStr string) (*AppClaims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(j.leeway),
	)
	token, err := parser.ParseWithClaims(tokenStr, &AppClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AppClaims); ok && token.Valid {
		// Optional checks: issuer, audience, not-before
		if j.issuer != "" && claims.Issuer != j.issuer {
			return nil, errors.New("invalid issuer")
		}
		if j.audience != "" && !containsAudience(claims.Audience, j.audience) {
			return nil, errors.New("invalid audience")
		}
		// NotBefore: if (now + leeway) is still before NBF, the token is not yet valid
		if claims.NotBefore != nil {
			now := time.Now()
			if now.Add(j.leeway).Before(claims.NotBefore.Time) {
				return nil, errors.New("token not yet valid")
			}
		}
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

func containsAudience(aud jwt.ClaimStrings, target string) bool {
	for _, a := range aud {
		if a == target {
			return true
		}
	}
	return false
}
