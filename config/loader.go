package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

type AuthConfiguration struct {
	Issuer          string   `json:"issuer"`
	JWKSUri         string   `json:"jwks_uri"`
	TokenSigningAlg []string `json:"id_token_signing_alg_values_supported"`
	Audience        string   `json:"audience"`
	ScopeClaim      string   `json:"scope_claim"`
	Scopes          []string `json:"scopes"`
	ClaimFields     []string `json:"claims"`
}

type Configuration struct {
	WebServerConfig *WebServerConfiguration
	AuthConfig      *AuthConfiguration
}

type WebServerConfiguration struct {
	Env              string
	Cors             cors.Cors
	EnableSwagger    bool
	WebPort          string
	TLSCertFile      string
	TLSCertKeyFile   string
	ConnectionString string
}

func loadAuthConfig() (*AuthConfiguration, error) {
	configUrl, ok := os.LookupEnv("AUTH_CONFIG_URL")
	if !ok {
		return nil, fmt.Errorf("AUTH_CONFIG_URL is a required parameter")
	}

	config, err := readOpenIdConfigurationFromURL(configUrl)

	if err != nil {
		return nil, err
	}

	config.Audience = os.Getenv("AUTH_AUDIENCE")

	scopeClaim, ok := os.LookupEnv("AUTH_SCOPE_CLAIM")
	if !ok {
		scopeClaim = "scp"
	}
	config.ScopeClaim = scopeClaim

	if scopes, ok := os.LookupEnv("AUTH_SCOPES"); ok {
		config.Scopes = strings.Split(scopes, ",")
	}

	if authClaims, ok := os.LookupEnv("AUTH_CLAIMS"); ok {
		config.ClaimFields = strings.Split(authClaims, ",")
	}

	return config, nil
}

func readOpenIdConfigurationFromURL(configUrl string) (*AuthConfiguration, error) {
	if configUrl == "" {
		return nil, fmt.Errorf("cannot read OpenId configuration without URL")
	}

	response, err := http.Get(configUrl)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	target := &AuthConfiguration{}
	err = json.NewDecoder(response.Body).Decode(target)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func loadWebServerConfig() (*WebServerConfiguration, error) {
	config := &WebServerConfiguration{}
	config.Env = os.Getenv("ENV")

	if err := godotenv.Load(); err != nil && config.Env == "" {
		return nil, err
	}

	allowedOrigin, _ := os.LookupEnv("ALLOWED_ORIGIN")

	config.Cors = *cors.New(cors.Options{
		AllowedOrigins: []string{allowedOrigin},
	})

	if enableSwagger, ok := os.LookupEnv("ENABLE_SWAGGER"); ok {
		config.EnableSwagger = enableSwagger == "true"
	}

	if webPort, ok := os.LookupEnv("WEB_PORT"); ok {
		config.WebPort = webPort
	} else {
		config.WebPort = "localhost:8000"
	}

	config.TLSCertFile, _ = os.LookupEnv("TLS_CERT_FILE")
	config.TLSCertKeyFile, _ = os.LookupEnv("TLS_CERT_KEY_FILE")

	if connectionString, ok := os.LookupEnv("DB_CONNECTION_STRING"); ok {
		config.ConnectionString = connectionString
	} else {
		return nil, fmt.Errorf("must set DB_CONNECTION_STRING=<connection string>")
	}

	return config, nil
}

func LoadConfig() *Configuration {
	webServerConfig, err := loadWebServerConfig()
	if err != nil {
		log.Fatal(err)
	}

	authConfig, err := loadAuthConfig()
	if err != nil {
		log.Fatal(err)
	}

	return &Configuration{
		WebServerConfig: webServerConfig,
		AuthConfig:      authConfig,
	}
}
