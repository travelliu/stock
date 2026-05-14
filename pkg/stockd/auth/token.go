package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

const TokenPrefix = "stk_"

// GenerateAPIToken returns (plainText, sha256Hex). The plain text is shown to
// the user once; only the hash is persisted.
func GenerateAPIToken() (string, string, error) {
	var buf [24]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", "", err
	}
	plain := TokenPrefix + base64.RawURLEncoding.EncodeToString(buf[:])
	return plain, HashToken(plain), nil
}

// HashToken returns the lowercase-hex sha256 of the plain token.
func HashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// ParseBearer returns the raw token portion of an Authorization header, or
// an error if the header is missing, malformed, or not a stockd token.
func ParseBearer(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("expected Bearer scheme")
	}
	tok := strings.TrimSpace(parts[1])
	if !strings.HasPrefix(tok, TokenPrefix) {
		return "", fmt.Errorf("token must start with %q", TokenPrefix)
	}
	return tok, nil
}
