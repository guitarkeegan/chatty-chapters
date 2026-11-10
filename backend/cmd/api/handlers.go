package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/google/uuid"
	"github.com/guitarkeegan/chatty-chapters/backend/internal/data"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {

	envel := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, envel, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) loggedIn(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, envelope{"logged_in": "true"}, nil)
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	app.clearAuthCookies(w)

	// TODO: remove from server store as well...
	app.writeJSON(w, http.StatusOK, envelope{"logout": map[string]string{"status": "true"}}, nil)
}

func (app *application) listStudentBooksHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := app.getUserIDFromSubContext(r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// get associated books
	books, err := app.models.UserBook.GetBooksByUserID(r.Context(), userID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	env := envelope{"books": books}

	app.writeJSON(w, http.StatusOK, env, nil)

}

// bookID is found from the listStudentBooksHandler
func (app *application) listChapters(w http.ResponseWriter, r *http.Request) {
	// context here?
	ctx := context.Background()

	userID, err := app.getUserIDFromSubContext(r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	bookID, err := app.readBookIDParam(r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	chapters, err := app.models.Chapter.GetAllVisibleChapters(ctx, userID, bookID) // get the course id
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	env := envelope{"chapters": chapters}

	app.writeJSON(w, http.StatusOK, env, nil)
}

func (app *application) listActivities(w http.ResponseWriter, r *http.Request) {

	ctx := context.Background()

	bookID, chapterID, err := app.readBookChapterParams(r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	userID, err := app.getUserIDFromSubContext(r)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// validate that use is enrolled in course with this book
	// get activities by chapter
	activities, err := app.models.UserActivity.GetActivitiesByUserBookChapter(ctx, userID, bookID, chapterID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	// if activitie don't exist, create them?
	if len(activities) == 0 {
		err := app.models.UserActivity.CreateActivitiesForUser(r.Context(), userID, bookID, []int64{chapterID})
		if err != nil {
			app.serverErrorResponse(w, r, fmt.Errorf("unable to create activities for userID: %d, bookID: %d, chapterIDs: %d", userID, bookID, chapterID))
		}
		activities, err = app.models.UserActivity.GetActivitiesByUserBookChapter(ctx, userID, bookID, chapterID)
		if err != nil {
			app.serverErrorResponse(w, r, errors.New("couldn't get these activities after creation"))
		}
	}

	env := envelope{"activities": activities}

	app.writeJSON(w, http.StatusOK, env, nil)
}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	dbg("LOGIN_HANDLER_START")
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &req)

	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("username and password required"))
		return
	}

	caser := cases.Lower(language.English)

	dbg("loginUser")
	resp, err := app.loginUser(caser.String(req.Username), req.Password)
	dbg("resp: %+v", resp)
	if err != nil {
		app.handleCognitoError(w, r, err)
		return
	}

	dbg("challenge check")
	if resp.ChallengeName != nil && *resp.ChallengeName == cognitoidentityprovider.ChallengeNameTypeNewPasswordRequired {
		dbg("  in challenge check")
		http.SetCookie(w, &http.Cookie{
			Name:     "music-session",
			Value:    *resp.Session,
			Path:     "/",
			HttpOnly: true,
			// TODO: for development
			// TODO: samesite
			// SameSite: http.SameSiteNoneMode, // for local
			Secure:  false,
			Expires: time.Now().Add(5 * time.Minute),
		})
		app.writeJSON(w, http.StatusOK, envelope{"challenge": map[string]string{"status": "NEW_PASSWORD_REQUIRED"}}, nil)
		return
	}

	dbg("LOGIN_HANDLER_END")

	app.sendIDAndRefreshToClient(w, http.StatusOK, envelope{
		"challenge": map[string]interface{}{"status": "SUCCESS"},
	}, resp.AuthenticationResult)

}

func (app *application) newPasswordHandler(w http.ResponseWriter, r *http.Request) {

	dbg("NEW_PASSWORD_HANDLER_START")

	var req struct {
		Username        string `json:"username"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	dbg("get cookie, 'music-session'")
	cookie, err := r.Cookie("music-session")
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("r.Cookie: %w", err))
		return
	}
	dbg("readJSON")
	app.readJSON(w, r, &req)

	dbg("handleNewPasswordChallenge")
	resp, err := app.handleNewPasswordChallenge(req.Username, req.Password, req.ConfirmPassword, cookie.Value)
	if err != nil {
		dbg("err not nil")
		if awsErr, ok := err.(awserr.Error); ok {
			dbg("at switch")
			switch awsErr.Code() {
			case cognitoidentityprovider.ErrCodeNotAuthorizedException:

				app.badRequestResponse(w, r, fmt.Errorf("not authorized"))
				return

			case cognitoidentityprovider.ErrCodeInvalidPasswordException:

				app.badRequestResponse(w, r, fmt.Errorf("invalid password: Password does not conform to policy. %s", awsErr.Message()))
				return

			case cognitoidentityprovider.ErrCodeUserNotFoundException:

				app.errorResponse(w, r, http.StatusNotFound, fmt.Errorf("user not found"))
				return

			case cognitoidentityprovider.ErrCodeLimitExceededException:

				app.errorResponse(w, r, http.StatusTooManyRequests, fmt.Errorf("too many reset attmepts, You'll have to wait to try again"))
				return
			default:
				// Handle other errors
				app.serverErrorResponse(w, r, fmt.Errorf("handleNewPasswordChallenge: %s", awsErr.Message()))
				return
			}
		} else {
			// Handle non-AWS errors
			dbg("else")
			app.serverErrorResponse(w, r, fmt.Errorf("handleNewPasswordChallenge: %w", err))
			return
		}
	}

	dbg("clear music-session")
	// Clear the session cookie after successful password change
	http.SetCookie(w, &http.Cookie{
		Name:     "music-session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // local dev
		Expires:  time.Now().Add(-1 * time.Hour),
	})

	dbg("NEW_PASSWORD_HANDLER_END")

	app.sendIDAndRefreshToClient(w, http.StatusOK, envelope{
		"challenge": map[string]interface{}{"status": "SUCCESS"},
	}, resp.AuthenticationResult)
}

// admin handler for registering a group of users
// course must already exist in the db
func (app *application) registerUsersAdminHandler(w http.ResponseWriter, r *http.Request) {

	var regList []data.UserRegistration
	// TODO: create middleware to check role
	// get an array of users
	err := app.readJSON(w, r, &regList)
	if err != nil {
		// should be admin facing only
		dbg("err on reading json")
		app.logger.Error("registerUsers", "error", err)
		app.badRequestResponse(w, r, fmt.Errorf("unable to decode"))
		return
	}
	// bulk insert
	err = app.models.User.InsertMany(&regList)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("insert Many: %w", err))
		return
	}

	app.writeJSON(w, http.StatusCreated, envelope{
		"register": map[string]interface{}{"status": "SUCCESS"},
	}, nil)

}

func (app *application) signupHandler(w http.ResponseWriter, r *http.Request) {

	dbg("START SIGNUP HANDLER")

	var req data.UserRegistration

	err := app.readJSON(w, r, &req)
	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("readJson: %w", err))
		return
	}

	if err := validateRegistrationInput(req.FirstName, req.LastName, req.Email); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	dbg("registering with cognito")
	// create user
	resp, err := app.registerCognitoUser(req.Email)
	if err != nil {
		app.logger.Error("registerCognitoUser", "msg", err.Error())
		app.errorResponse(w, r, http.StatusUnauthorized, "not authorized")
		return
	}
	dbg("cognito resp: %+v", resp)

	dbg("get user sub")
	userSub, err := app.getUserSub(req.Email)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("signupHandler: getUserSub: %w", err))
		return
	}
	dbg("parse userSub")
	req.CognitoSub, err = uuid.Parse(userSub)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("signupHandler: uuid.Parse: %w", err))
		return
	}

	dbg("registering user in app")
	if err = app.models.User.InsertPublic(&req); err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("signupHandler: InsertPublic: %w", err))
		return
	}

	// TODO: would need to delete the user that was just created in cognito...

	dbg("SIGNUP HANDLER END")

	app.writeJSON(w, http.StatusCreated, envelope{
		"register": map[string]interface{}{"status": "SUCCESS", "message": "A temporary password has been send to your email"},
	}, nil)

}

// username and email will be the same with the app and cognito
func (app *application) registerCognitoUser(username string) (*cognitoidentityprovider.AdminCreateUserOutput, error) {

	input := &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId: aws.String(app.auth.userPoolID),
		Username:   aws.String(username),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(username),
			},
			{
				Name:  aws.String("email_verified"),
				Value: aws.String("false"),
			},
		},
	}

	resp, err := app.auth.cognitoClient.AdminCreateUser(input)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return resp, nil

}

func (app *application) getUserSub(username string) (string, error) {
	input := &cognitoidentityprovider.AdminGetUserInput{
		UserPoolId: aws.String(app.auth.userPoolID),
		Username:   aws.String(username),
	}

	resp, err := app.auth.cognitoClient.AdminGetUser(input)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Find the sub attribute
	for _, attr := range resp.UserAttributes {
		if *attr.Name == "sub" {
			return *attr.Value, nil
		}
	}

	return "", fmt.Errorf("sub attribute not found for user: %s", username)
}

func (app *application) userProgressHandler(w http.ResponseWriter, r *http.Request) {

	dbg("userProgressHandler")

	userID, err := app.getUserIDFromSubContext(r)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("userProgresssHandler: getid from sub: %w", err))
		return
	}

	// get all courses associated with the user
	// pass in request context
	courses, err := app.models.UserCourse.GetAll(userID)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("userProgressHandler: GetAll: %w", err))
	}

	dbg("userProgressHandler: End")

	app.writeJSON(w, http.StatusOK, envelope{"user_progress": courses}, nil)

}

func (app *application) userCourseProgressHandler(w http.ResponseWriter, r *http.Request) {

	dbg("userCourseProgressHandler start")

	userID, err := app.getUserIDFromSubContext(r)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("userProgresssHandler: getid from sub: %w", err))
		return
	}

	bookID, err := app.readBookIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// get course stats!!
	userProgress, err := app.models.UserProgress.Get(r.Context(), userID, bookID)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("userCourseProgresssHandler: Get: %w", err))
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"user_progress": userProgress}, nil)
}

func (app *application) forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {

	// get the body
	dbg("FORGOT_PASS_HANDLER")
	var req struct {
		Username         string `json:"username"`
		Password         string `json:"password"`
		ConfirmationCode string `json:"confirmation_code"`
	}

	err := app.readJSON(w, r, &req)

	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("username and password required"))
		return
	}

	caser := cases.Lower(language.English)

	confirmForgotPasswordInput := &cognitoidentityprovider.ConfirmForgotPasswordInput{
		ClientId:         aws.String(app.auth.clientID),
		Username:         aws.String(caser.String(req.Username)),
		ConfirmationCode: aws.String(req.ConfirmationCode), // The code user received
		Password:         aws.String(req.Password),
	}

	_, err = app.auth.cognitoClient.ConfirmForgotPassword(confirmForgotPasswordInput)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case cognitoidentityprovider.ErrCodeNotAuthorizedException:

				app.badRequestResponse(w, r, fmt.Errorf("not authorized"))
				return

			case cognitoidentityprovider.ErrCodeInvalidPasswordException:

				app.badRequestResponse(w, r, fmt.Errorf("invalid password: Password does not conform to policy. %s", awsErr.Message()))
				return

			case cognitoidentityprovider.ErrCodeUserNotFoundException:

				app.errorResponse(w, r, http.StatusNotFound, fmt.Errorf("user Not Found"))
				return

			case cognitoidentityprovider.ErrCodeLimitExceededException:

				app.errorResponse(w, r, http.StatusTooManyRequests, fmt.Errorf("too many reset attmepts, You'll have to wait to try again"))
				return
			default:
				// Handle other errors
				app.serverErrorResponse(w, r, fmt.Errorf("confirmForgotPassword: %s", awsErr.Message()))
				return
			}
		} else {
			// Handle non-AWS errors
			app.serverErrorResponse(w, r, fmt.Errorf("confirmForgotPassword: %w", err))
			return
		}
	} else {
		app.writeJSON(w, http.StatusOK, envelope{"status": "SUCCESS", "message": "return to login"}, nil)
	}

}

func (app *application) resetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	dbg("RESET_PASS_HANDLER")

	var req struct {
		Username string `json:"username"`
	}

	err := app.readJSON(w, r, &req)

	if err != nil {
		app.badRequestResponse(w, r, fmt.Errorf("username required"))
		return
	}

	caser := cases.Lower(language.English)

	forgotPasswordInput := &cognitoidentityprovider.ForgotPasswordInput{
		ClientId: aws.String(app.auth.clientID),
		Username: aws.String(caser.String(req.Username)),
	}

	_, err = app.auth.cognitoClient.ForgotPassword(forgotPasswordInput)
	if err != nil {
		app.serverErrorResponse(w, r, fmt.Errorf("forgotPassword: %w", err))
	} else {
		// Instructions have been sent to the user to reset their password
		app.writeJSON(w, http.StatusOK, envelope{"status": "SUCCESS", "message": "check your email for a confirmation code"}, nil)
	}
}
