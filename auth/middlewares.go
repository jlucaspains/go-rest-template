package auth

import (
	"context"
	"encoding/json"
	"goapi-template/middlewares"
	"goapi-template/models"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	keyfunc "github.com/MicahParks/keyfunc/v2"
	"github.com/open-policy-agent/opa/rego"
)

var authConfig *models.AuthConfiguration
var opaQuery *rego.PreparedEvalQuery
var cachedSet JKWS

func Init() {
	authConfig = loadConfig()
	opaQuery = loadOpaQuery()
	cachedSet = loadJWKSCache()
}

type key int

const UserKey key = 1

func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		returnResult := "valid"

		defer func() {
			elapsed := time.Since(start)
			slog.Debug("Auth middleware",
				"timeElapsed", elapsed,
				"result", returnResult,
				"traceId", r.Context().Value(middlewares.ContextKey("traceId")))
		}()

		token, err := extractToken(r)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			data := &models.ErrorResult{Errors: []string{"Auth token was not provided or is invalid"}}
			result, _ := json.Marshal(data)
			w.Write(result)
			returnResult = err.Error()
			return
		}

		user, err := validateUserToken(token, authConfig, cachedSet)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			data := &models.ErrorResult{Errors: []string{"Auth token is invalid"}}
			result, _ := json.Marshal(data)
			w.Write(result)
			returnResult = err.Error()
			return
		}

		newReq := r.WithContext(context.WithValue(r.Context(), UserKey, user))

		next.ServeHTTP(w, newReq)
	})
}

func OpaMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		returnResult := "allowed"

		defer func() {
			elapsed := time.Since(start)
			slog.Debug("OPA middleware",
				"timeElapsed", elapsed,
				"result", returnResult,
				"traceId", r.Context().Value(middlewares.ContextKey("traceId")))
		}()

		token, _ := extractToken(r)

		input := map[string]interface{}{
			"method": r.Method,
			"path":   r.RequestURI,
			"token":  token,
		}
		res, err := opaQuery.Eval(r.Context(), rego.EvalInput(input))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			result, _ := json.Marshal(err)
			w.Write(result)
			returnResult = err.Error()
			return
		}

		if !res.Allowed() {
			w.WriteHeader(http.StatusForbidden)
			w.Header().Set("Content-Type", "application/json")
			result, _ := json.Marshal(&models.ErrorResult{Errors: []string{"forbidden"}})
			w.Write(result)
			returnResult = "forbidden"
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loadConfig() *models.AuthConfiguration {
	configUrl, ok := os.LookupEnv("AUTH_CONFIG_URL")
	if !ok {
		log.Fatal("AUTH_CONFIG_URL is a required parameter")
	}

	config, err := readAuthConfiguration(configUrl)
	if err != nil {
		log.Fatalf("Failed to load Auth configuration. Error: %v", err)
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

	return config
}

func loadOpaQuery() *rego.PreparedEvalQuery {
	regoPath, ok := os.LookupEnv("AUTH_REGO_PATH")

	if !ok {
		regoPath = "./auth/authz.rego"
	}

	query, err := rego.New(rego.Query("data.authz.allow"), rego.Load([]string{regoPath}, nil)).PrepareForEval(context.TODO())
	if err != nil {
		log.Fatalf("failed to create rego query. Error: %v", err)
	}

	return &query
}

func loadJWKSCache() *keyfunc.JWKS {
	options := keyfunc.Options{
		RefreshInterval: time.Hour,
		RefreshTimeout:  time.Second * 10,
		RefreshErrorHandler: func(err error) {
			slog.Error("There was an error with the jwt.Keyfunc", "error", err.Error())
		},
	}

	jwks, err := keyfunc.Get(authConfig.JWKSUri, options)
	if err != nil {
		log.Fatalf("Failed to create JWKS from resource at the given URL.\nError: %s", err.Error())
	}

	return jwks
}
