package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/guitarkeegan/music-class-tutor/internal/ai"
	"github.com/guitarkeegan/music-class-tutor/internal/data"
	"github.com/joho/godotenv"
)

func mockAuthApp() (*application, error) {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error on env load: %q", err)
	}

	cognitoRegion := os.Getenv("AWS_REGION")
	clientID := os.Getenv("USER_POOL_CLIENT_ID")
	userPoolID := os.Getenv("USER_POOL_ID")

	return &application{
		config: config{},
		logger: &slog.Logger{},
		models: data.Models{},
		ai:     &ai.Ai{},
		auth: &authProvider{
			cognitoClient: initializeCognito(),
			userPoolID:    userPoolID,
			clientID:      clientID,
			cognitoRegion: cognitoRegion,
			cognitoJwkURL: fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", cognitoRegion, userPoolID),
		},
	}, nil

}

func TestGetPublicKeys(t *testing.T) {
	app, err := mockAuthApp()
	if err != nil {
		t.Errorf("error on app setup: %s", err)
	}

	got, err := app.getPublicKeys()
	if err != nil {
		t.Errorf("error getting signing key: %s", err)
	}

	fmt.Printf("got something! %+v", got)

}
