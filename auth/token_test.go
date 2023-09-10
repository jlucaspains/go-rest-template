package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"goapi-template/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func generateRsaKeyPair() (*rsa.PrivateKey, *rsa.PublicKey) {
	privkey, _ := rsa.GenerateKey(rand.Reader, 4096)
	return privkey, &privkey.PublicKey
}

type TestJWKS struct {
	PublicKey *rsa.PublicKey
}

func (j TestJWKS) Keyfunc(token *jwt.Token) (interface{}, error) {
	return j.PublicKey, nil
}

func TestValidToken(t *testing.T) {
	authConfig := &models.AuthConfiguration{
		Issuer:          "issuer",
		TokenSigningAlg: []string{"RS256"},
		Audience:        "audience",
		ScopeClaim:      "scp",
		Scopes:          []string{"api"},
		ClaimFields:     []string{"test"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"aud":   "audience",
		"scp":   "api",
		"sub":   "sub",
		"name":  "name",
		"email": "test@email.com",
		"test":  "test",
		"nbf":   time.Now().Add(1).Unix(),
	})

	priv, pub := generateRsaKeyPair()

	jwks := &TestJWKS{PublicKey: pub}

	// Sign and get the complete encoded token as a string using the secret
	tokenString, _ := token.SignedString(priv)

	user, err := validateUserToken(tokenString, authConfig, jwks)

	assert.Nil(t, err)
	assert.Equal(t, "test@email.com", user.Email)
	assert.Contains(t, user.Claims, "test")
}

func TestTokenExpired(t *testing.T) {
	authConfig := &models.AuthConfiguration{
		Issuer:          "issuer",
		TokenSigningAlg: []string{"RS256"},
		Audience:        "audience",
		ScopeClaim:      "scp",
		Scopes:          []string{"api"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"aud":   "audience",
		"scp":   "api",
		"sub":   "sub",
		"name":  "name",
		"email": "test@email.com",
		"nbf":   time.Now().Add(time.Hour * -2).Unix(),
		"exp":   time.Now().Add(time.Hour * -1).Unix(),
	})

	priv, pub := generateRsaKeyPair()

	jwks := &TestJWKS{PublicKey: pub}

	// Sign and get the complete encoded token as a string using the secret
	tokenString, _ := token.SignedString(priv)

	user, err := validateUserToken(tokenString, authConfig, jwks)

	assert.NotNil(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "token has invalid claims: token is expired", err.Error())
}

func TestBadTokenAudience(t *testing.T) {
	authConfig := &models.AuthConfiguration{
		Issuer:          "issuer",
		TokenSigningAlg: []string{"RS256"},
		Audience:        "audience",
		ScopeClaim:      "scp",
		Scopes:          []string{"api"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"aud":   "nope",
		"scp":   "api",
		"sub":   "sub",
		"name":  "name",
		"email": "test@email.com",
		"nbf":   time.Now().Add(time.Hour * -1).Unix(),
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
	})

	priv, pub := generateRsaKeyPair()

	jwks := &TestJWKS{PublicKey: pub}

	// Sign and get the complete encoded token as a string using the secret
	tokenString, _ := token.SignedString(priv)

	user, err := validateUserToken(tokenString, authConfig, jwks)

	assert.NotNil(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "token not issue to correct audience", err.Error())
}

func TestBadTokenScope(t *testing.T) {
	authConfig := &models.AuthConfiguration{
		Issuer:          "issuer",
		TokenSigningAlg: []string{"RS256"},
		Audience:        "audience",
		ScopeClaim:      "scp",
		Scopes:          []string{"api"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"aud":   "audience",
		"scp":   "nope",
		"sub":   "sub",
		"name":  "name",
		"email": "test@email.com",
		"nbf":   time.Now().Add(time.Hour * -1).Unix(),
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
	})

	priv, pub := generateRsaKeyPair()

	jwks := &TestJWKS{PublicKey: pub}

	// Sign and get the complete encoded token as a string using the secret
	tokenString, _ := token.SignedString(priv)

	user, err := validateUserToken(tokenString, authConfig, jwks)

	assert.NotNil(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "token doesn't have valid scopes", err.Error())
}

func TestBadTokenSignature(t *testing.T) {
	authConfig := &models.AuthConfiguration{
		Issuer:          "issuer",
		TokenSigningAlg: []string{"RS256"},
		Audience:        "audience",
		ScopeClaim:      "scp",
		Scopes:          []string{"api"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"aud":   "audience",
		"scp":   "api",
		"sub":   "sub",
		"name":  "name",
		"email": "test@email.com",
		"nbf":   time.Now().Add(time.Hour * -1).Unix(),
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
	})

	priv, _ := generateRsaKeyPair()
	_, pub := generateRsaKeyPair() // to fail validation

	jwks := &TestJWKS{PublicKey: pub}

	// Sign and get the complete encoded token as a string using the secret
	tokenString, _ := token.SignedString(priv)

	user, err := validateUserToken(tokenString, authConfig, jwks)

	assert.NotNil(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "token signature is invalid: crypto/rsa: verification error", err.Error())
}

func TestBadTokenSignatureAlg(t *testing.T) {
	authConfig := &models.AuthConfiguration{
		Issuer:          "issuer",
		TokenSigningAlg: []string{"RS256"},
		Audience:        "audience",
		ScopeClaim:      "scp",
		Scopes:          []string{"api"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"aud":   "audience",
		"scp":   "api",
		"sub":   "sub",
		"name":  "name",
		"email": "test@email.com",
		"nbf":   time.Now().Add(time.Hour * -1).Unix(),
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
	})

	priv, pub := generateRsaKeyPair()

	jwks := &TestJWKS{PublicKey: pub}

	// Sign and get the complete encoded token as a string using the secret
	tokenString, _ := token.SignedString(priv)

	user, err := validateUserToken(tokenString, authConfig, jwks)

	assert.NotNil(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "token signature alg and issuer alg do not match", err.Error())
}

func TestLoadAuthConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"issuer":"issuer",
			"jwks_uri":"jwks_uri",
			"id_token_signing_alg_values_supported":["alg"]
		}`))
	}))
	defer server.Close()

	os.Setenv("AUTH_SCOPES", "api,test")
	os.Setenv("AUTH_CLAIMS", "sid")
	os.Setenv("AUTH_CONFIG_URL", server.URL)
	os.Setenv("AUTH_AUDIENCE", "aud")

	config := loadConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "issuer", config.Issuer)
	assert.Equal(t, "jwks_uri", config.JWKSUri)
	assert.Contains(t, config.TokenSigningAlg, "alg")
}

func TestLoadJWKSCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"keys": [
				{
					"kty": "RSA",
					"use": "sig",
					"kid": "nOo3ZDrODXEK1jKWhXslHR_KXEg",
					"x5t": "nOo3ZDrODXEK1jKWhXslHR_KXEg",
					"n": "oaLLT9hkcSj2tGfZsjbu7Xz1Krs0qEicXPmEsJKOBQHauZ_kRM1HdEkgOJbUznUspE6xOuOSXjlzErqBxXAu4SCvcvVOCYG2v9G3-uIrLF5dstD0sYHBo1VomtKxzF90Vslrkn6rNQgUGIWgvuQTxm1uRklYFPEcTIRw0LnYknzJ06GC9ljKR617wABVrZNkBuDgQKj37qcyxoaxIGdxEcmVFZXJyrxDgdXh9owRmZn6LIJlGjZ9m59emfuwnBnsIQG7DirJwe9SXrLXnexRQWqyzCdkYaOqkpKrsjuxUj2-MHX31FqsdpJJsOAvYXGOYBKJRjhGrGdONVrZdUdTBQ",
					"e": "AQAB",
					"x5c": [
						"MIIDBTCCAe2gAwIBAgIQN33ROaIJ6bJBWDCxtmJEbjANBgkqhkiG9w0BAQsFADAtMSswKQYDVQQDEyJhY2NvdW50cy5hY2Nlc3Njb250cm9sLndpbmRvd3MubmV0MB4XDTIwMTIyMTIwNTAxN1oXDTI1MTIyMDIwNTAxN1owLTErMCkGA1UEAxMiYWNjb3VudHMuYWNjZXNzY29udHJvbC53aW5kb3dzLm5ldDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKGiy0/YZHEo9rRn2bI27u189Sq7NKhInFz5hLCSjgUB2rmf5ETNR3RJIDiW1M51LKROsTrjkl45cxK6gcVwLuEgr3L1TgmBtr/Rt/riKyxeXbLQ9LGBwaNVaJrSscxfdFbJa5J+qzUIFBiFoL7kE8ZtbkZJWBTxHEyEcNC52JJ8ydOhgvZYykete8AAVa2TZAbg4ECo9+6nMsaGsSBncRHJlRWVycq8Q4HV4faMEZmZ+iyCZRo2fZufXpn7sJwZ7CEBuw4qycHvUl6y153sUUFqsswnZGGjqpKSq7I7sVI9vjB199RarHaSSbDgL2FxjmASiUY4RqxnTjVa2XVHUwUCAwEAAaMhMB8wHQYDVR0OBBYEFI5mN5ftHloEDVNoIa8sQs7kJAeTMA0GCSqGSIb3DQEBCwUAA4IBAQBnaGnojxNgnV4+TCPZ9br4ox1nRn9tzY8b5pwKTW2McJTe0yEvrHyaItK8KbmeKJOBvASf+QwHkp+F2BAXzRiTl4Z+gNFQULPzsQWpmKlz6fIWhc7ksgpTkMK6AaTbwWYTfmpKnQw/KJm/6rboLDWYyKFpQcStu67RZ+aRvQz68Ev2ga5JsXlcOJ3gP/lE5WC1S0rjfabzdMOGP8qZQhXk4wBOgtFBaisDnbjV5pcIrjRPlhoCxvKgC/290nZ9/DLBH3TbHk8xwHXeBAnAjyAqOZij92uksAv7ZLq4MODcnQshVINXwsYshG1pQqOLwMertNaY5WtrubMRku44Dw7R"
					],
					"issuer": "https://localhost/"
				}
			]
		}`))
	}))
	defer server.Close()

	os.Setenv("AUTH_CONFIG_URL", server.URL)
	os.Setenv("AUTH_AUDIENCE", "aud")

	authConfig = &models.AuthConfiguration{
		JWKSUri: server.URL,
	}
	cache := loadJWKSCache()

	assert.NotNil(t, cache)
	assert.Contains(t, cache.KIDs(), "nOo3ZDrODXEK1jKWhXslHR_KXEg")
}

func TestLoadOPAQuery(t *testing.T) {
	query := loadOpaQuery("./authz.rego")

	assert.NotNil(t, query)
}

func TestAuthTokenMiddlewareWithoutToken(t *testing.T) {
	router := mux.NewRouter()
	router.Use(TokenAuthMiddleware())

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}

	router.HandleFunc("/test", handler).Methods("GET")

	reqFound, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, reqFound)

	assert.Equal(t, 401, w.Code)
}

func TestAuthMiddlewareValid(t *testing.T) {
	authConfig = &models.AuthConfiguration{
		Issuer:          "issuer",
		TokenSigningAlg: []string{"RS256"},
		Audience:        "audience",
		ScopeClaim:      "scp",
		Scopes:          []string{"api"},
		ClaimFields:     []string{"test"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"aud":   "audience",
		"scp":   "api",
		"sub":   "sub",
		"name":  "name",
		"email": "test@email.com",
		"test":  "test",
		"nbf":   time.Now().Add(1).Unix(),
	})

	priv, pub := generateRsaKeyPair()
	cachedSet = &TestJWKS{PublicKey: pub}

	// Sign and get the complete encoded token as a string using the secret
	tokenString, _ := token.SignedString(priv)

	router := mux.NewRouter()
	router.Use(TokenAuthMiddleware())

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}

	router.HandleFunc("/test", handler).Methods("GET")

	reqFound, _ := http.NewRequest("GET", "/test", nil)
	reqFound.Header.Add("Authorization", fmt.Sprintf("Bearer %v", tokenString))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, reqFound)

	assert.Equal(t, 200, w.Code)
}

func TestOPAMiddlewareValid(t *testing.T) {
	opaQuery = loadOpaQuery("./test.rego")

	router := mux.NewRouter()
	router.Use(OpaMiddleware())

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}

	router.HandleFunc("/test", handler).Methods("GET")

	reqFound, _ := http.NewRequest("GET", "/test", nil)
	reqFound.Header.Add("Authorization", fmt.Sprintf("Bearer %v", "pass"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, reqFound)

	assert.Equal(t, 200, w.Code)
}

func TestOPAMiddleware403(t *testing.T) {
	opaQuery = loadOpaQuery("./test.rego")

	router := mux.NewRouter()
	router.Use(OpaMiddleware())

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}

	router.HandleFunc("/test", handler).Methods("GET")

	reqFound, _ := http.NewRequest("GET", "/test", nil)
	reqFound.Header.Add("Authorization", fmt.Sprintf("Bearer %v", "deny"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, reqFound)

	assert.Equal(t, 403, w.Code)
}
