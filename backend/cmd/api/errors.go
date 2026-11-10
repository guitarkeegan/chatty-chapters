package main

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}

// for internal 500 errors, log and send back generic response
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {

	app.logError(r, err)

	msg := "The server encountered a problem and could not process your request."
	app.errorResponse(w, r, http.StatusInternalServerError, msg)

}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	msg := "The requested resource cannot be found."
	app.errorResponse(w, r, http.StatusNotFound, msg)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	app.logger.Error("failed to authenticate", "error", err.Error())

	message := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *application) handleCognitoError(w http.ResponseWriter, r *http.Request, err error) {
	var errorMessage string
	var statusCode int

	if awsErr, ok := err.(awserr.Error); ok {

		// AWS Cognito returns specific error codes, we match them here
		switch awsErr.Code() {
		case cognitoidentityprovider.ErrCodeNotAuthorizedException:
			errorMessage = "Incorrect username or password."
			statusCode = http.StatusUnauthorized
		case cognitoidentityprovider.ErrCodeUserNotFoundException:
			errorMessage = "User does not exist."
			statusCode = http.StatusUnauthorized
		case cognitoidentityprovider.ErrCodeUserNotConfirmedException:
			errorMessage = "User account not confirmed. Please check your email."
			statusCode = http.StatusForbidden
		case cognitoidentityprovider.ErrCodePasswordResetRequiredException:
			errorMessage = "Password reset required. Please reset your password."
			statusCode = http.StatusForbidden
		case cognitoidentityprovider.ErrCodeInvalidPasswordException:
			errorMessage = "The password you provided is invalid."
			statusCode = http.StatusBadRequest
		case cognitoidentityprovider.ErrCodeLimitExceededException:
			errorMessage = "Too many attempts. Please try again later."
			statusCode = http.StatusTooManyRequests
		case cognitoidentityprovider.ErrCodeTooManyFailedAttemptsException:
			errorMessage = "Too many attempts. Please try again later."
			statusCode = http.StatusTooManyRequests
		case cognitoidentityprovider.ErrCodeUsernameExistsException:
			errorMessage = "Problem with login. Please contact instructor"
			statusCode = http.StatusBadRequest
		default:
			// If it's not a known Cognito error, return nil so the calling function can handle it
		}
	}
	if errorMessage == "" {
		errorMessage = "Unknown Error, please contact your instructor"
		statusCode = http.StatusInternalServerError
	}

	// Send the specific error response to the user
	app.errorResponse(w, r, statusCode, errorMessage)
}
