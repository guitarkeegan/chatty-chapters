package main

import (
	"context"
	"errors"
	"expvar"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lestrrat-go/jwx/jwk"

	"github.com/guitarkeegan/chatty-chapters/backend/internal/ai"
	"github.com/guitarkeegan/chatty-chapters/backend/internal/data"
	"github.com/guitarkeegan/chatty-chapters/backend/internal/seed"
	"github.com/joho/godotenv"
)

var dbg = func() func(format string, as ...any) {
	if os.Getenv("DEBUG") == "" {
		return func(string, ...any) {}
	}
	return func(format string, as ...any) {
		log.Printf(format, as...)
	}
}()

type jwkCache struct {
	expiresAt time.Time
	jwkSet    *jwk.Set
}

type authProvider struct {
	cognitoClient *cognitoidentityprovider.CognitoIdentityProvider
	userPoolID    string
	clientID      string
	cognitoRegion string
	cognitoJwkURL string
	jwkCache      *jwkCache
	// TODO: create seperate structure for caching here. time create. expires at. anything else...
}

func initializeCognito(region string) *cognitoidentityprovider.CognitoIdentityProvider {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		log.Fatalf("Failed to initialize AWS session: %v", err)
	}

	cognitoClient := cognitoidentityprovider.New(sess)
	if cognitoClient == nil {
		log.Fatalf("cognitoClient is nil")
	}

	return cognitoClient
}

// envOr returns the value of the environment variable key, or fallback if it is unset
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

var version = "1.0.0"
var wsConnectionCount int32

type config struct {
	port         int
	env          string
	cookieDomain string
	db           struct {
		dsn string
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	logger *slog.Logger
	models data.Models
	ai     *ai.Ai
	auth   *authProvider
}

func main() {

	expvar.NewString("version").Set(version)
	// Publish the number of active goroutines.
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	// Publish the current Unix timestamp.
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))
	var cognitoRegion string
	var cfg config
	var seedEmbeddings bool
	var seedAll bool
	var clearAllUserProgress bool
	var userPoolID string
	var clientID string

	// a .env file is optional, variables may come from the environment instead
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatalf("error on env load: %q", err)
	}

	flag.IntVar(&cfg.port, "port", 8080, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "development|staging|production")
	flag.StringVar(&cfg.cookieDomain, "cookie-domain", os.Getenv("COOKIE_DOMAIN"), "domain for auth cookies in production, e.g. example.com")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("DNS_CONN_DEV"), "db connect string")
	flag.BoolVar(&seedEmbeddings, "seed_embeddings", false, "seed db with book embeddings")
	flag.BoolVar(&seedAll, "seed_all", false, "seed db with users, books, chapters, user-activities")
	flag.BoolVar(&clearAllUserProgress, "clear_all_user_progress", false, "clear all user related data")
	flag.StringVar(&cognitoRegion, "cognito_region", envOr("AWS_REGION", "us-west-1"), "set aws cognito region")
	flag.StringVar(&userPoolID, "user_pool_id", os.Getenv("USER_POOL_ID"), "set user pool id")
	flag.StringVar(&clientID, "cognito_client_id", os.Getenv("USER_POOL_CLIENT_ID"), "set user pool client id")
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	var cognitoJwkURL = fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", cognitoRegion, userPoolID)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	fmt.Println(cfg.cors.trustedOrigins)

	cognitoClient := initializeCognito(cognitoRegion)

	auth := &authProvider{
		cognitoClient: cognitoClient,
		userPoolID:    userPoolID,
		clientID:      clientID,
		cognitoRegion: cognitoRegion,
		cognitoJwkURL: cognitoJwkURL,
		jwkCache:      &jwkCache{},
	}

	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// Publish the database connection pool statistics.
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stat()
	}))

	if seedEmbeddings {
		err := seed.SeedBookEmbeddings(db, os.Getenv("OPENAI"))
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}

	if seedAll {
		err := seed.SeedUsersBooksChaptersRelations(db)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}

	if clearAllUserProgress {
		err := seed.ClearAllUserProgress(context.Background(), db)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}

	// the seed and clear flags are one off tasks, so don't start the server
	if seedEmbeddings || seedAll || clearAllUserProgress {
		return
	}

	ai, err := ai.New(os.Getenv("OPENAI"), cfg.env, db)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		ai:     ai,
		auth:   auth,
	}
	app.logger.Info("llms ready", "env", cfg.env)
	app.logger.Info("database connection pool established")

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	app.logger.Info("Server started", "addr", srv.Addr, "env", cfg.env)

	err = srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(cfg config) (*pgxpool.Pool, error) {

	// TODO: add config for DB
	db, err := pgxpool.New(context.Background(), cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Return the sql.DB connection pool.
	return db, nil
}
