package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwk"
)

type Jwks struct {
	Keys []json.RawMessage `json:"keys"`
}

type MyClaims struct {
	jwt.StandardClaims
}

func (jwkCache *jwkCache) get() *jwk.Set {
	// if now greater than expiresAt, return nil
	if time.Now().After(jwkCache.expiresAt) {
		return nil
	}
	return jwkCache.jwkSet

}

func (app *application) loginUser(username, password string) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	dbg("loginUser")
	authParams := map[string]*string{
		"USERNAME": aws.String(username),
		"PASSWORD": aws.String(password),
	}

	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow:       aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: authParams,
		// TODO: pass in
		ClientId: aws.String(app.auth.clientID),
	}

	dbg("loginUser return")
	return app.auth.cognitoClient.InitiateAuth(input)
}

func (app *application) handleNewPasswordChallenge(username, tempPassword, newPassword, session string) (*cognitoidentityprovider.RespondToAuthChallengeOutput, error) {

	dbg("handleNewPasswordChallenge start")
	challengeResponses := map[string]*string{
		"USERNAME":     aws.String(username),
		"PASSWORD":     aws.String(tempPassword),
		"NEW_PASSWORD": aws.String(newPassword),
	}

	input := &cognitoidentityprovider.RespondToAuthChallengeInput{
		ChallengeName:      aws.String("NEW_PASSWORD_REQUIRED"),
		ChallengeResponses: challengeResponses,
		ClientId:           aws.String(app.auth.clientID),
		Session:            aws.String(session),
	}

	// id, access, and refresh
	res, err := app.auth.cognitoClient.RespondToAuthChallenge(input)
	if err != nil {
		return nil, err
	}

	dbg("handleNewPasswordChallenge return")
	return res, nil

}

// verify token and return the user's sub and experation
// TODO: incorperate refetching logic in the case that the verification fails the first time?
func (app *application) verifyToken(jwtToken string) (*MyClaims, error) {

	dbg("verifyToken start")
	var jwks *jwk.Set
	// add the cache check here
	jwksCached := app.auth.jwkCache.get()
	if jwksCached != nil {
		jwks = jwksCached
	} else {
		var err error
		jwks, err = app.getPublicKeys()
		if err != nil {
			return nil, fmt.Errorf("error getting public keys: %v", err)
		}
	}
	// calls the cognito endpoint, parses the jwk set
	// returns a jwk.Set, which is an interface for dealing with jwk sets
	// basically an array with 2 jwk objects
	// TODO: this can be cached and updated when neccissary
	// Parse the token with validation
	// the header has the alg and kid only
	dbg("parse token")
	parsedToken, err := jwt.ParseWithClaims(jwtToken, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {

		// TODO: remove this to get a raise in 6 months
		// time.Sleep(300 * time.Millisecond)
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the kid from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid token header: kid not found")
		}

		// Find the corresponding JWK
		jwkKey, err := findJWK(*jwks, kid)
		if err != nil {
			return nil, fmt.Errorf("error finding JWK: %v", err)
		}

		// Get the RSA public key
		var publicKey rsa.PublicKey
		if err := jwkKey.Raw(&publicKey); err != nil {
			return nil, fmt.Errorf("error converting JWK to RSA public key: %v", err)
		}

		return &publicKey, nil

	})

	if err != nil {
		fmt.Printf("error during token parsing: %v", err)
		return nil, fmt.Errorf("error verifying token: %v", err)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	// Convert claims to MyClaims
	// TODO: is this neccissary?
	claims, ok := parsedToken.Claims.(*MyClaims)
	if !ok {
		return nil, fmt.Errorf("token claims are not of type MyClaims")
	}

	// TODO: horrible things with nbf...
	return claims, nil

}

func (app *application) getPublicKeys() (*jwk.Set, error) {
	dbg("getPublicKeys start")
	// check cache first here

	resp, err := http.Get(app.auth.cognitoJwkURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get public keys: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	keySet, err := jwk.Parse(body)
	if err != nil {
		return nil, err
	}
	dbg("getPulicKeys return")
	return &keySet, nil

}

// verify that the jwt is from cognito by
// checking that the jwkSet includes the kid
func findJWK(jwkSet jwk.Set, kid string) (jwk.Key, error) {

	for it := jwkSet.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		key := pair.Value.(jwk.Key)
		if key.KeyID() == kid {
			return key, nil
		}
	}
	return nil, fmt.Errorf("key not found")

}

func (app *application) refreshToken(refreshToken string) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	input := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: aws.String("REFRESH_TOKEN_AUTH"),
		ClientId: aws.String(app.auth.clientID),
		AuthParameters: map[string]*string{
			"REFRESH_TOKEN": aws.String(refreshToken),
		},
	}

	return app.auth.cognitoClient.InitiateAuth(input)
}
