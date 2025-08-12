package security

import (
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
	GenerateToken(userID string, role string) (string, error)
	ValidateToken(tokenStr string) (*AppClaims, error)
}

type jwtService struct {
	// common
	expireDuration time.Duration
	issuer         string
	audience       string
	leeway         time.Duration
	// signing algorithm and key id
	alg string // HS256 (default), RS256, EDDSA
	kid string
	// HS256 secret
	hsSecret string
	// RS256 keys
	rsPrivate *rsa.PrivateKey
	rsPublics map[string]*rsa.PublicKey // kid -> key
	// EdDSA keys
	edPrivate ed25519.PrivateKey
	edPublics map[string]ed25519.PublicKey // kid -> key
}

func NewJWTService(secret string, expireSec int) JWTService {
	return &jwtService{
		alg:            "HS256",
		hsSecret:       secret,
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

	var method jwt.SigningMethod
	switch strings.ToUpper(j.alg) {
	case "HS256":
		method = jwt.SigningMethodHS256
	case "RS256":
		method = jwt.SigningMethodRS256
	case "EDDSA":
		method = jwt.SigningMethodEdDSA
	default:
		return "", fmt.Errorf("unsupported jwt alg: %s", j.alg)
	}
	token := jwt.NewWithClaims(method, claims)
	// Always set type header
	token.Header["typ"] = "JWT"
	if j.kid != "" {
		token.Header["kid"] = j.kid
	}
	switch strings.ToUpper(j.alg) {
	case "HS256":
		return token.SignedString([]byte(j.hsSecret))
	case "RS256":
		if j.rsPrivate == nil {
			return "", errors.New("missing RS256 private key")
		}
		return token.SignedString(j.rsPrivate)
	case "EDDSA":
		if j.edPrivate == nil {
			return "", errors.New("missing EdDSA private key")
		}
		return token.SignedString(j.edPrivate)
	}
	return "", errors.New("unreachable")
}

func (j *jwtService) ValidateToken(tokenStr string) (*AppClaims, error) {
	var methods []string
	switch strings.ToUpper(j.alg) {
	case "HS256":
		methods = []string{jwt.SigningMethodHS256.Alg()}
	case "RS256":
		methods = []string{jwt.SigningMethodRS256.Alg()}
	case "EDDSA":
		methods = []string{jwt.SigningMethodEdDSA.Alg()}
	default:
		return nil, fmt.Errorf("unsupported jwt alg: %s", j.alg)
	}
	parser := jwt.NewParser(
		jwt.WithValidMethods(methods),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(j.leeway),
	)
	token, err := parser.ParseWithClaims(tokenStr, &AppClaims{}, func(token *jwt.Token) (interface{}, error) {
		switch strings.ToUpper(j.alg) {
		case "HS256":
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(j.hsSecret), nil
		case "RS256":
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("unexpected signing method")
			}
			kid, _ := token.Header["kid"].(string)
			if kid != "" {
				if pk, ok := j.rsPublics[kid]; ok {
					return pk, nil
				}
				return nil, fmt.Errorf("unknown kid: %s", kid)
			}
			switch len(j.rsPublics) {
			case 0:
				return nil, errors.New("no RS256 public keys configured")
			case 1:
				for _, pk := range j.rsPublics { // return the only key
					return pk, nil
				}
			default:
				return nil, errors.New("kid required when multiple RS256 public keys are configured")
			}
			return nil, errors.New("unreachable")
		case "EDDSA":
			if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, errors.New("unexpected signing method")
			}
			kid, _ := token.Header["kid"].(string)
			if kid != "" {
				if pk, ok := j.edPublics[kid]; ok {
					return pk, nil
				}
				return nil, fmt.Errorf("unknown kid: %s", kid)
			}
			switch len(j.edPublics) {
			case 0:
				return nil, errors.New("no EdDSA public keys configured")
			case 1:
				for _, pk := range j.edPublics {
					return pk, nil
				}
			default:
				return nil, errors.New("kid required when multiple EdDSA public keys are configured")
			}
			return nil, errors.New("unreachable")
		}
		return nil, errors.New("unreachable")
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

// ConfigureAlgorithm sets algorithm, kid, and keys (for RS256/EdDSA). For HS256, ensure hsSecret is set.
func (j *jwtService) ConfigureAlgorithm(alg, kid, privateKeyPath, privateKeyPEM, publicKeysDir string) error {
	j.alg = strings.ToUpper(strings.TrimSpace(alg))
	j.kid = strings.TrimSpace(kid)
	switch j.alg {
	case "HS256":
		if j.hsSecret == "" {
			return errors.New("JWT_SECRET required for HS256")
		}
		return nil
	case "RS256":
		pk, err := loadRSAPrivateKey(privateKeyPath, privateKeyPEM)
		if err != nil {
			return err
		}
		j.rsPrivate = pk
		pubs, err := loadRSAPublicKeys(publicKeysDir)
		if err != nil {
			return err
		}
		j.rsPublics = pubs
		return nil
	case "EDDSA":
		edPriv, err := loadEdPrivateKey(privateKeyPath, privateKeyPEM)
		if err != nil {
			return err
		}
		j.edPrivate = edPriv
		pubs, err := loadEdPublicKeys(publicKeysDir)
		if err != nil {
			return err
		}
		j.edPublics = pubs
		return nil
	default:
		return fmt.Errorf("unsupported jwt alg: %s", alg)
	}
}

func loadFileIfExists(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	return os.ReadFile(path)
}

func loadRSAPrivateKey(path, pemStr string) (*rsa.PrivateKey, error) {
	var data []byte
	if pemStr != "" {
		data = []byte(pemStr)
	} else {
		b, err := loadFileIfExists(path)
		if err != nil {
			return nil, err
		}
		data = b
	}
	if len(data) == 0 {
		return nil, errors.New("empty RSA private key")
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid RSA private key PEM")
	}
	var key any
	var err error
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported RSA key type: %s", block.Type)
	}
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}
	return rsaKey, nil
}

func loadRSAPublicKeys(dir string) (map[string]*rsa.PublicKey, error) {
	out := make(map[string]*rsa.PublicKey)
	if dir == "" {
		return out, nil
	}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		block, _ := pem.Decode(b)
		if block == nil {
			return fmt.Errorf("invalid PEM in %s", path)
		}
		var pub any
		switch block.Type {
		case "PUBLIC KEY":
			pub, err = x509.ParsePKIXPublicKey(block.Bytes)
		case "RSA PUBLIC KEY":
			pub, err = x509.ParsePKCS1PublicKey(block.Bytes)
		default:
			return fmt.Errorf("unsupported public key type in %s: %s", path, block.Type)
		}
		if err != nil {
			return err
		}
		switch k := pub.(type) {
		case *rsa.PublicKey:
			out[name] = k
		default:
			return fmt.Errorf("non-RSA public key in %s", path)
		}
		return nil
	})
	return out, err
}

func loadEdPrivateKey(path, pemStr string) (ed25519.PrivateKey, error) {
	var data []byte
	if pemStr != "" {
		data = []byte(pemStr)
	} else {
		b, err := loadFileIfExists(path)
		if err != nil {
			return nil, err
		}
		data = b
	}
	if len(data) == 0 {
		return nil, errors.New("empty Ed private key")
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("invalid Ed private key PEM")
	}
	var key any
	var err error
	switch block.Type {
	case "PRIVATE KEY":
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported Ed key type: %s", block.Type)
	}
	if err != nil {
		return nil, err
	}
	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("not an Ed25519 private key")
	}
	return edKey, nil
}

func loadEdPublicKeys(dir string) (map[string]ed25519.PublicKey, error) {
	out := make(map[string]ed25519.PublicKey)
	if dir == "" {
		return out, nil
	}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		block, _ := pem.Decode(b)
		if block == nil {
			return fmt.Errorf("invalid PEM in %s", path)
		}
		if block.Type != "PUBLIC KEY" {
			return fmt.Errorf("unsupported public key type in %s: %s", path, block.Type)
		}
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return err
		}
		k, ok := pub.(ed25519.PublicKey)
		if !ok {
			return fmt.Errorf("non-Ed25519 public key in %s", path)
		}
		out[name] = k
		return nil
	})
	return out, err
}
