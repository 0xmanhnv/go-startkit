package security

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
    GenerateToken(userID string, role string) (string, error)
    ValidateToken(tokenStr string) (*jwt.RegisteredClaims, error)
}

type jwtService struct {
    secretKey string
    expireDuration time.Duration
}

func NewJWTService(secret string, expireSec int) JWTService {
    return &jwtService{
        secretKey: secret,
        expireDuration: time.Duration(expireSec) * time.Second,
    }
}

func (j *jwtService) GenerateToken(userID string, role string) (string, error) {
    claims := jwt.RegisteredClaims{
        Subject:   userID,
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expireDuration)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(j.secretKey))
}

func (j *jwtService) ValidateToken(tokenStr string) (*jwt.RegisteredClaims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(j.secretKey), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
        return claims, nil
    }
    return nil, errors.New("invalid token")
}