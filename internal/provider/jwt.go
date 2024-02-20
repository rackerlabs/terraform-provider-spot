package provider

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RxtSpotToken struct {
	token       string
	parsedToken *jwt.Token
	claims      jwt.MapClaims
}

func NewRxtSpotToken(token string) *RxtSpotToken {
	return &RxtSpotToken{token: token}
}

func (j *RxtSpotToken) Parse() error {
	var err error
	j.parsedToken, _, err = jwt.NewParser(jwt.WithExpirationRequired()).ParseUnverified(j.token, &j.claims)
	return err
}

func (j *RxtSpotToken) GetOrgID() (string, error) {
	if val, found := j.claims["org_id"]; found {
		if orgID, ok := val.(string); ok {
			return orgID, nil
		} else {
			return "", errors.New("org_id is not of string type")
		}
	} else {
		return "", errors.New("org_id not found")
	}
}

func (j *RxtSpotToken) IsExpired() (bool, error) {
	if exp, err := j.claims.GetExpirationTime(); err != nil {
		return false, fmt.Errorf("failed to get expiration time: %w", err)
	} else {
		if exp.Time.Before(time.Now().UTC()) {
			return true, nil
		}
	}
	return false, nil
}

func (j *RxtSpotToken) IsEmailVerified() bool {
	if val, found := j.claims["email_verified"]; found {
		if emailVerified, ok := val.(bool); ok {
			if !emailVerified {
				return false
			}
		}
	}
	// If email_verified is not present or not a boolean, we assume it is verified
	return true
}

func (j *RxtSpotToken) IsValidSignature() (bool, error) {
	var err error
	var issuer string
	if issuer, err = j.claims.GetIssuer(); err != nil {
		return false, fmt.Errorf("failed to get issuer: %w", err)
	}
	var ok bool
	var kid string
	if kidIface, found := j.parsedToken.Header["kid"]; !found {
		return false, errors.New("kid not found in token header")
	} else if kid, ok = kidIface.(string); !ok {
		return false, errors.New("kid from token header is not a string")
	}

	// jwks URL: https://cloudspaces.us.auth0.com/.well-known/jwks.json
	jwksURL := fmt.Sprintf("%s.well-known/jwks.json", issuer)
	publicKey, err := j.getIssuerPublicKey(jwksURL, kid)
	if err != nil {
		return false, fmt.Errorf("error getting public key from issuer: %v", err)
	}

	_, err = jwt.Parse(j.token, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if err != nil {
		return false, fmt.Errorf("error verifying JWT signature: %v", err)
	}

	return true, nil
}

func fromBase64URL(s string) ([]byte, error) {
	// Replace '-' with '+' and '_' with '/' to convert base64url to base64
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")

	// Add padding if needed
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}

	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (j *RxtSpotToken) getIssuerPublicKey(url string, kid string) (*rsa.PublicKey, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching public keys from issuer: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Parse JSON response to extract public keys
	var jwks struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, err
	}

	// Find the RSA public key
	for _, key := range jwks.Keys {
		if key.Kty == "RSA" && key.Kid == kid {
			n, err := fromBase64URL(key.N)
			if err != nil {
				return nil, fmt.Errorf("error decoding public key modulus: %v", err)
			}
			e, err := fromBase64URL(key.E)
			if err != nil {
				return nil, fmt.Errorf("error decoding public key exponent: %v", err)
			}
			return &rsa.PublicKey{
				N: new(big.Int).SetBytes(n),
				E: int(new(big.Int).SetBytes(e).Int64()),
			}, nil
		}
	}
	return nil, errors.New("public key not found")
}
