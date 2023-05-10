package auth

import (
	"context"
	"goapi-template/models"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	keyfunc "github.com/MicahParks/keyfunc/v2"
	"github.com/gin-gonic/gin"
	"github.com/open-policy-agent/opa/rego"
)

var authConfig *models.AuthConfiguration
var opaQuery *rego.PreparedEvalQuery
var cachedSet JKWS

func Init() {
	authConfig = loadConfig()
	opaQuery = loadOpaQuery("./auth/authz.rego")
	cachedSet = loadJWKSCache()
}

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		token, err := extractToken(c.Request)

		if err != nil {
			log.Printf("token check failed %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, &models.ErrorResult{Errors: []string{"Auth token was not provided or is invalid"}})
			return
		}

		user, err := validateUserToken(token, authConfig, cachedSet)

		if err != nil {
			log.Printf("token check failed %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, &models.ErrorResult{Errors: []string{"Auth token was not provided or is invalid"}})
			return
		}

		c.Set("User", user)

		elapsed := time.Since(start)
		log.Printf("Auth Middleware took %v", elapsed)

		c.Next()
	}
}

func OpaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		token, _ := extractToken(c.Request)

		input := map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.RequestURI,
			"token":  token,
		}
		res, err := opaQuery.Eval(c.Request.Context(), rego.EvalInput(input))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			c.Abort()
			return
		}

		if !res.Allowed() {
			c.JSON(http.StatusForbidden, &models.ErrorResult{Errors: []string{"forbidden"}})
			c.Abort()
			return
		}

		elapsed := time.Since(start)
		log.Printf("Opa Middleware took %v", elapsed)

		c.Next()
	}
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

func loadOpaQuery(regoPath string) *rego.PreparedEvalQuery {
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
			log.Printf("There was an error with the jwt.Keyfunc\nError: %s", err.Error())
		},
	}

	jwks, err := keyfunc.Get(authConfig.JWKSUri, options)
	if err != nil {
		log.Fatalf("Failed to create JWKS from resource at the given URL.\nError: %s", err.Error())
	}

	return jwks
}
