package config

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadWebConfig(t *testing.T) {
	t.Setenv("ENV", "TEST")
	t.Setenv("ALLOWED_ORIGIN", "localhost:8000")
	t.Setenv("ENABLE_SWAGGER", "true")
	t.Setenv("WEB_PORT", "localhost:8000")
	t.Setenv("TLS_CERT_FILE", "tls_cert_file")
	t.Setenv("TLS_CERT_KEY_FILE", "tls_cert_key_file")
	t.Setenv("DB_CONNECTION_STRING", "connection_string")

	config, err := loadWebServerConfig()

	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "localhost:8000", config.WebPort)
	assert.Equal(t, "connection_string", config.ConnectionString)
	assert.True(t, config.EnableSwagger)
	assert.Contains(t, "TEST", config.Env)
	assert.Contains(t, "tls_cert_file", config.TLSCertFile)
	assert.Contains(t, "tls_cert_key_file", config.TLSCertKeyFile)
}

func TestLoadWebConfigDefaults(t *testing.T) {
	t.Setenv("ENV", "TEST")
	t.Setenv("DB_CONNECTION_STRING", "connection_string")

	config, err := loadWebServerConfig()

	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "localhost:8000", config.WebPort)
	assert.Equal(t, "connection_string", config.ConnectionString)
	assert.False(t, config.EnableSwagger)
	assert.Contains(t, "TEST", config.Env)
	assert.Empty(t, config.TLSCertFile)
	assert.Empty(t, config.TLSCertKeyFile)
}

func TestLoadWebConfigMissingConnectionString(t *testing.T) {
	t.Setenv("ENV", "TEST")

	_, err := loadWebServerConfig()

	assert.Error(t, err, "must set DB_CONNECTION_STRING=<connection string>")
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

	t.Setenv("AUTH_SCOPES", "api,test")
	t.Setenv("AUTH_CLAIMS", "sid")
	t.Setenv("AUTH_CONFIG_URL", server.URL)
	t.Setenv("AUTH_AUDIENCE", "aud")

	config, err := loadAuthConfig()

	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "issuer", config.Issuer)
	assert.Equal(t, "jwks_uri", config.JWKSUri)
	assert.Contains(t, config.TokenSigningAlg, "alg")
}

func TestLoadAuthConfigMissingUrl(t *testing.T) {
	_, err := loadAuthConfig()

	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "AUTH_CONFIG_URL is a required parameter")
}

func TestLoadAuthConfigBadUrl(t *testing.T) {
	t.Setenv("AUTH_CONFIG_URL", "http://localhost")

	_, err := loadAuthConfig()

	assert.NotNil(t, err)
}

func TestLoadCacheConfigDisableCache(t *testing.T) {
	t.Setenv("ENABLE_TRANSPARENT_CACHE", "false")

	config, _ := loadCacheConfig()

	assert.NotNil(t, config)
	assert.False(t, config.EnableTransparentCaching)
}

func TestLoadCacheConfigValidDefaults(t *testing.T) {
	t.Setenv("ENABLE_TRANSPARENT_CACHE", "true")

	config, _ := loadCacheConfig()

	assert.NotNil(t, config)
	assert.True(t, config.EnableTransparentCaching)
	assert.Equal(t, "localhost:6379", config.RedisAddress)
	assert.Equal(t, "", config.RedisPassword)
	assert.Equal(t, time.Hour, config.Expiration)
	assert.Equal(t, 0, config.RedisDb)
}

func TestLoadCacheConfigValid(t *testing.T) {
	t.Setenv("ENABLE_TRANSPARENT_CACHE", "true")
	t.Setenv("REDIS_ADDRESS", "localhost:6379")
	t.Setenv("REDIS_PASSWORD", "password")
	t.Setenv("REDIS_DB", "1")
	t.Setenv("REDIS_DEFAULT_EXPIRATION", "2h")

	config, _ := loadCacheConfig()

	assert.NotNil(t, config)
	assert.True(t, config.EnableTransparentCaching)
	assert.Equal(t, "localhost:6379", config.RedisAddress)
	assert.Equal(t, "password", config.RedisPassword)
	assert.Equal(t, time.Hour*2, config.Expiration)
	assert.Equal(t, 1, config.RedisDb)
}

func TestLoadAllConfig(t *testing.T) {
	t.Setenv("ENV", "TEST")
	t.Setenv("ALLOWED_ORIGIN", "localhost:8000")
	t.Setenv("ENABLE_SWAGGER", "true")
	t.Setenv("WEB_PORT", "localhost:8000")
	t.Setenv("TLS_CERT_FILE", "tls_cert_file")
	t.Setenv("TLS_CERT_KEY_FILE", "tls_cert_key_file")
	t.Setenv("DB_CONNECTION_STRING", "connection_string")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"issuer":"issuer",
			"jwks_uri":"jwks_uri",
			"id_token_signing_alg_values_supported":["alg"]
		}`))
	}))
	defer server.Close()

	t.Setenv("AUTH_SCOPES", "api,test")
	t.Setenv("AUTH_CLAIMS", "sid")
	t.Setenv("AUTH_CONFIG_URL", server.URL)
	t.Setenv("AUTH_AUDIENCE", "aud")
	t.Setenv("ENABLE_TRANSPARENT_CACHE", "false")

	config := LoadConfig()

	assert.NotNil(t, config)
}
