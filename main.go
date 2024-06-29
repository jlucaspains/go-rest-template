// gorest-template REST API
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"goapi-template/auth"
	"goapi-template/db"
	"goapi-template/docs"
	"goapi-template/handlers"
	"goapi-template/middlewares"
)

type configuration struct {
	env              string
	cors             cors.Cors
	enableSwagger    bool
	webPort          string
	tlsCertFile      string
	tlsCertKeyFile   string
	connectionString string
}

var config configuration

func initConfiguration() {
	config.env = os.Getenv("ENV")

	if err := godotenv.Load(); err != nil && config.env == "" {
		log.Fatal(fmt.Printf("Error loading .env file: %s", err))
	}

	allowedOrigin, _ := os.LookupEnv("ALLOWED_ORIGIN")

	config.cors = *cors.New(cors.Options{
		AllowedOrigins: []string{allowedOrigin},
	})

	if enableSwagger, ok := os.LookupEnv("ENABLE_SWAGGER"); ok {
		config.enableSwagger = enableSwagger == "true"
	}

	if webPort, ok := os.LookupEnv("WEB_PORT"); ok {
		config.webPort = webPort
	} else {
		config.webPort = "localhost:8000"
	}

	config.tlsCertFile, _ = os.LookupEnv("TLS_CERT_FILE")
	config.tlsCertKeyFile, _ = os.LookupEnv("TLS_CERT_KEY_FILE")

	if connectionString, ok := os.LookupEnv("DB_CONNECTION_STRING"); ok {
		config.connectionString = connectionString
	} else {
		log.Fatal("must set DB_CONNECTION_STRING=<connection string>")
	}
}

func withMiddlewares(handler func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return middlewares.LogMiddleware(
		config.cors.Handler(
			auth.OpaMiddleware(
				auth.TokenAuthMiddleware(
					http.HandlerFunc(handler)))))
}

func onlyLogMiddleware(handler func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return middlewares.LogMiddleware(
		http.HandlerFunc(handler))
}

func setupRouter(db db.Querier) http.Handler {
	slog.Info("Starting API... \n")

	controllers := &handlers.Handlers{Queries: db}
	router := http.NewServeMux()

	router.HandleFunc("OPTIONS /", config.cors.HandlerFunc)
	router.Handle("GET /health", onlyLogMiddleware(controllers.GetHealth))

	router.Handle("GET /person/{id}", withMiddlewares(controllers.GetPerson))
	router.Handle("POST /person", withMiddlewares(controllers.PostPerson))
	router.Handle("PUT /person/{id}", withMiddlewares(controllers.PutPerson))
	router.Handle("DELETE /person/{id}", withMiddlewares(controllers.DeletePerson))

	if config.enableSwagger {
		slog.Info("Swagger enabled")
		swaggerHandler := httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
			httpSwagger.DeepLinking(true),
			httpSwagger.DocExpansion("none"),
			httpSwagger.DomID("swagger-ui"),
		)
		router.Handle("GET /swagger/", swaggerHandler)
	}

	return router
}

func initDB(ctx context.Context) (db.Querier, func()) {
	if err := db.Init(config.connectionString); err != nil {
		log.Fatal(err)
	}

	conn, err := pgxpool.New(ctx, config.connectionString)
	if err != nil {
		log.Fatal(err)
	}

	queries := db.New(conn)

	return queries, conn.Close
}

func startWebServer(querier db.Querier) func(ctx context.Context) error {
	slog.Info("Setting up API router...\n")
	docs.SwaggerInfo.BasePath = "/"

	router := setupRouter(querier)

	srv := &http.Server{
		Addr: config.webPort,
	}
	srv.Handler = router

	useTls := config.tlsCertFile != "" && config.tlsCertKeyFile != ""

	slog.Info("Starting TLS server", "port", config.webPort, "tls", useTls)

	var err error
	if useTls {
		err = srv.ListenAndServeTLS(config.tlsCertFile, config.tlsCertKeyFile)
	} else {
		err = srv.ListenAndServe()
	}

	if err != nil {
		slog.Error("Error starting server", "error", err)
	}

	return srv.Shutdown
}

// @securitydefinitions.oauth2.implicit					OAuth2Implicit
// @authorizationUrl										https://login.microsoftonline.com/9e6b9f31-c202-4cbd-a9b1-7e5cb3874384/oauth2/v2.0/authorize
// @tokenUrl												https://login.microsoftonline.com/9e6b9f31-c202-4cbd-a9b1-7e5cb3874384/oauth2/v2.0/token
// @scope.api://c571ab3c-0fde-43b2-b010-77e7bdd0d6f7/api	API
func main() {
	ctx := context.Background()

	slog.Info("loading .env file...\n")
	initConfiguration()

	slog.Info("Init auth...\n")
	auth.Init()

	slog.Info("Init DB...\n")
	db, dbDispose := initDB(ctx)
	defer dbDispose()

	webDispose := startWebServer(db)
	defer webDispose(ctx)
}
