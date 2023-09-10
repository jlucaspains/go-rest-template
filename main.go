// gorest-template REST API
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/gorm"

	goHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"goapi-template/auth"
	"goapi-template/db"
	"goapi-template/docs"
	"goapi-template/handlers"
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

func setupRouter(db *gorm.DB) http.Handler {
	log.Print("Starting API... \n")

	handlers := &handlers.Handlers{DB: db}

	router := mux.NewRouter()
	peopleRouter := router.PathPrefix("/person").Subrouter()
	peopleRouter.HandleFunc("/{id}", handlers.GetPerson).Methods("GET")
	peopleRouter.HandleFunc("", handlers.PostPerson).Methods("POST")
	peopleRouter.HandleFunc("/{id}", handlers.PutPerson).Methods("PUT")
	peopleRouter.HandleFunc("/{id}", handlers.DeletePerson).Methods("DELETE")
	peopleRouter.Use(auth.TokenAuthMiddleware())
	peopleRouter.Use(auth.OpaMiddleware())

	router.HandleFunc("/health", handlers.GetHealth).Methods("GET")

	headersOk := goHandlers.AllowedHeaders([]string{"X-Requested-With", "Origin", "Content-Length", "Content-Type"})
	originsOk := goHandlers.AllowedOrigins([]string{getAllowedOrigins()})
	methodsOk := goHandlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	credentialsOk := goHandlers.AllowCredentials()

	// logMiddleware := midlewares.NewLogMiddleware(log.Default())
	// router.Use(logMiddleware.Func())

	enableSwagger, ok := os.LookupEnv("ENABLE_SWAGGER")
	if !ok {
		enableSwagger = "false"
	}

	if enableSwagger == "true" {
		router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"), //The url pointing to API definition
			httpSwagger.DeepLinking(true),
			httpSwagger.DocExpansion("none"),
			httpSwagger.DomID("swagger-ui"),
		)).Methods(http.MethodGet)
	}

	handler := goHandlers.CORS(originsOk, headersOk, methodsOk, credentialsOk)(router)

	return goHandlers.LoggingHandler(os.Stdout, handler)
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

	useTls := false
	certFile, ok := os.LookupEnv("TLS_CERT_FILE")
	useTls = ok

	certKeyFile, ok := os.LookupEnv("TLS_CERT_KEY_FILE")
	useTls = useTls && ok

	log.Printf("Starting TLS server on port: %s; use tls: %t", port, useTls)
	if useTls {
		log.Fatalln(http.ListenAndServeTLS(port, certFile, certKeyFile, router))
	} else {
		log.Fatalln(http.ListenAndServe(port, router))
	}
}
