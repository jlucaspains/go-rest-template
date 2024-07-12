// gorest-template REST API
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"goapi-template/auth"
	"goapi-template/config"
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

var configValues *config.Configuration

func withMiddlewares(handler func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return middlewares.TraceMiddleware(
		middlewares.LogMiddleware(
			configValues.WebServerConfig.Cors.Handler(
				auth.TokenAuthMiddleware(
					auth.OpaMiddleware(
						http.HandlerFunc(handler))))))
}

func onlyLogMiddleware(handler func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return middlewares.TraceMiddleware(
		middlewares.LogMiddleware(
			http.HandlerFunc(handler)))
}

func setupRouter(db db.Querier) http.Handler {
	slog.Info("Starting API... \n")

	controllers := handlers.New(db)
	router := http.NewServeMux()

	router.HandleFunc("OPTIONS /", configValues.WebServerConfig.Cors.HandlerFunc)
	router.Handle("GET /health", onlyLogMiddleware(controllers.GetHealth))

	router.Handle("GET /person/{id}", withMiddlewares(controllers.GetPerson))
	router.Handle("POST /person", withMiddlewares(controllers.PostPerson))
	router.Handle("PUT /person/{id}", withMiddlewares(controllers.PutPerson))
	router.Handle("DELETE /person/{id}", withMiddlewares(controllers.DeletePerson))

	if configValues.WebServerConfig.EnableSwagger {
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

func initDB(ctx context.Context, configValues *config.Configuration) (db.Querier, func()) {
	if err := db.Init(configValues.WebServerConfig.ConnectionString); err != nil {
		log.Fatal(err)
	}

	conn, err := pgxpool.New(ctx, configValues.WebServerConfig.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	queries := db.New(conn)

	return queries, conn.Close
}

func startWebServer(querier db.Querier, configValues *config.Configuration) func(ctx context.Context) error {
	slog.Info("Setting up API router...\n")
	docs.SwaggerInfo.BasePath = "/"

	router := setupRouter(querier)

	srv := &http.Server{
		Addr: configValues.WebServerConfig.WebPort,
	}
	srv.Handler = router

	useTls := configValues.WebServerConfig.TLSCertFile != "" && configValues.WebServerConfig.TLSCertKeyFile != ""

	slog.Info("Starting TLS server", "port", configValues.WebServerConfig.WebPort, "tls", useTls)

	var err error
	if useTls {
		err = srv.ListenAndServeTLS(configValues.WebServerConfig.TLSCertFile, configValues.WebServerConfig.TLSCertKeyFile)
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
	configValues = config.LoadConfig()

	slog.Info("Init auth...\n")
	auth.Init(configValues.AuthConfig)

	slog.Info("Init DB...\n")
	db, dbDispose := initDB(ctx, configValues)
	defer dbDispose()

	webDispose := startWebServer(db, configValues)
	defer webDispose(ctx)
}
