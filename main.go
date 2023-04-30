// gorest-template REST API
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"goapi-template/auth"
	"goapi-template/db"
	"goapi-template/docs"
	"goapi-template/handlers"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func loadEnv() {
	// outside of local environment, variables should be
	// OS environment variables
	env := os.Getenv("ENV")
	if err := godotenv.Load(); err != nil && env == "" {
		log.Fatal(fmt.Printf("Error loading .env file: %s", err))
	}
}

func getAllowedOrigins() string {
	allowedOrigin, ok := os.LookupEnv("ALLOWED_ORIGIN")
	if !ok {
		allowedOrigin = "http://localhost:8000"
	}

	return allowedOrigin
}

func setupRouter(db *gorm.DB) *gin.Engine {
	log.Print("Starting API... \n")

	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{getAllowedOrigins()},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	setTrustedProxies(engine)

	handlers := &handlers.Handlers{DB: db}

	authGroup := engine.Group("person", auth.TokenAuthMiddleware(), auth.OpaMiddleware())
	{
		authGroup.GET(":id", handlers.GetPerson)
		authGroup.POST("", handlers.PostPerson)
		authGroup.PUT(":id", handlers.PutPerson)
		authGroup.DELETE(":id", handlers.DeletePerson)
	}

	engine.GET("/health", handlers.GetHealth)

	enableSwagger, ok := os.LookupEnv("ENABLE_SWAGGER")
	if !ok {
		enableSwagger = "false"
	}

	if enableSwagger == "true" {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler, ginSwagger.Oauth2DefaultClientID("c571ab3c-0fde-43b2-b010-77e7bdd0d6f7")))
	}

	return engine
}

func initDB() *gorm.DB {
	provider, ok := os.LookupEnv("DB_PROVIDER")
	if !ok {
		provider = "postgres"
	}

	connectionString, ok := os.LookupEnv("DB_CONNECTION_STRING")
	if !ok {
		log.Fatal("DB_CONNECTION_STRING is a required parameter")
	}

	db, err := db.Init(provider, connectionString, true)

	if err != nil {
		log.Fatalf("Failed to initialize DB. Error: %v", err)
	}

	return db
}

func setTrustedProxies(engine *gin.Engine) {
	if config, ok := os.LookupEnv("GIN_TRUSTED_PROXIES"); ok {
		if config == "nil" {
			engine.SetTrustedProxies(nil)
		} else {
			result := strings.Split(config, ",")
			engine.SetTrustedProxies(result)
		}
	}
}

// @securitydefinitions.oauth2.implicit					OAuth2Implicit
// @authorizationUrl										https://login.microsoftonline.com/9e6b9f31-c202-4cbd-a9b1-7e5cb3874384/oauth2/v2.0/authorize
// @tokenUrl												https://login.microsoftonline.com/9e6b9f31-c202-4cbd-a9b1-7e5cb3874384/oauth2/v2.0/token
// @scope.api://c571ab3c-0fde-43b2-b010-77e7bdd0d6f7/api	API
func main() {
	log.Print("loading .env file...\n")
	loadEnv()

	log.Print("Init auth...\n")
	auth.Init()

	log.Print("Init DB...\n")
	db := initDB()

	log.Print("Setting up API router...\n")
	docs.SwaggerInfo.BasePath = "/"

	router := setupRouter(db)

	port, ok := os.LookupEnv("WEB_PORT")
	if !ok {
		port = "localhost:8000"
	}
	router.Run(port)
}
