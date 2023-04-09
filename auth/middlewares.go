package auth

import (
	"context"
	"goapi-template/models"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/open-policy-agent/opa/rego"
)

var authConfig *models.AuthConfiguration
var opaQuery *rego.PreparedEvalQuery

func Init() {
	authConfig = loadConfig()
	opaQuery = loadOpaQuery()
}

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := isTokenValid(c.Request, authConfig)

		if err != nil {
			log.Printf("token check failed %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, &models.ErrorResult{Errors: []string{"Auth token was not provided or is invalid"}})
			return
		}

		c.Next()
	}
}

func OpaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := extractToken(c.Request)

		input := map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.RequestURI,
			"token":  token,
			"subject": map[string]interface{}{
				"user":  "user",
				"group": "groups",
			},
		}
		res, err := opaQuery.Eval(context.TODO(), rego.EvalInput(input))
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

	return config
}

func loadOpaQuery() *rego.PreparedEvalQuery {
	query, err := rego.New(rego.Query("data.authz.allow"), rego.Load([]string{"./auth/authz.rego"}, nil)).PrepareForEval(context.TODO())
	if err != nil {
		log.Fatalf("failed to create rego query. Error: %v", err)
	}

	return &query
}
