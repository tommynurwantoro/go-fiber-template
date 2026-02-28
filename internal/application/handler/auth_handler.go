package handler

import (
	"app/internal/adapter/email"
	"app/internal/adapter/oauth"
	"app/internal/application/model"
	"app/internal/application/service"
	"app/internal/domain"
	"app/internal/pkg/formatter"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/tommynurwantoro/golog"
)

type AuthHandler interface {
	Register(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	Logout(c *fiber.Ctx) error
	RefreshTokens(c *fiber.Ctx) error
	ForgotPassword(c *fiber.Ctx) error
	ResetPassword(c *fiber.Ctx) error
	SendVerificationEmail(c *fiber.Ctx) error
	VerifyEmail(c *fiber.Ctx) error
	GoogleLogin(c *fiber.Ctx) error
	GoogleCallback(c *fiber.Ctx) error
}

type AuthHandlerImpl struct {
	AuthService   service.AuthService  `inject:"authService"`
	EmailAdapter  email.EmailAdapter   `inject:"email"`
	GoogleAdapter oauth.GoogleAdapter  `inject:"oauth"`
	TokenService  service.TokenService `inject:"tokenService"`
	UserService   service.UserService  `inject:"userService"`
}

// @Tags         Auth
// @Summary      Register as user
// @Description  Create a new user account. Returns user data and auth tokens on success.
// @Accept       json
// @Produce      json
// @Param        request  body  model.RegisterRequest  true  "Request body (name, email, password with strong-password validation)"
// @Router       /auth/register [post]
// @Success      201  {object}  model.RegisterResponse
// @Failure      400  {object}  model.ErrorInvalidRequest  "Invalid request body or validation failed"
// @Failure      409  {object}  model.ErrorDuplicateEmail  "Email already taken"
func (a *AuthHandlerImpl) Register(c *fiber.Ctx) error {
	req := new(model.RegisterRequest)

	if err := c.BodyParser(req); err != nil {
		golog.Error("Error parsing request body", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := a.AuthService.Register(c, req)
	if err != nil {
		return err
	}

	accessToken, refreshToken, err := a.TokenService.GenerateAuthTokens(c, user.ID.String())
	if err != nil {
		return err
	}

	resp := &model.RegisterResponse{
		ID:                    user.ID.String(),
		Name:                  user.Name,
		Email:                 user.Email,
		AccessToken:           accessToken.Token,
		AccessTokenExpiresAt:  accessToken.Expires,
		RefreshToken:          refreshToken.Token,
		RefreshTokenExpiresAt: refreshToken.Expires,
	}

	return c.Status(fiber.StatusCreated).JSON(formatter.NewSuccessResponse(formatter.Success, "Register successfully", resp))
}

// @Tags         Auth
// @Summary      Login
// @Description  Authenticate with email and password. Returns access and refresh tokens on success.
// @Accept       json
// @Produce      json
// @Param        request  body  model.LoginRequest  true  "Request body (email, password with strong-password validation)"
// @Router       /auth/login [post]
// @Success      200  {object}  model.LoginResponse
// @Failure      400  {object}  model.ErrorInvalidRequest  "Invalid request body or validation failed"
// @Failure      401  {object}  model.ErrorFailedLogin  "Invalid email or password"
// @Failure      404  {object}  model.ErrorNotFound  "User not found"
func (a *AuthHandlerImpl) Login(c *fiber.Ctx) error {
	req := new(model.LoginRequest)

	if err := c.BodyParser(req); err != nil {
		golog.Error("Error parsing request body", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	user, err := a.AuthService.Login(c, req)
	if err != nil {
		return err
	}

	accessToken, refreshToken, err := a.TokenService.GenerateAuthTokens(c, user.ID.String())
	if err != nil {
		return err
	}

	resp := &model.LoginResponse{
		AccessToken:           accessToken.Token,
		AccessTokenExpiresAt:  accessToken.Expires,
		RefreshToken:          refreshToken.Token,
		RefreshTokenExpiresAt: refreshToken.Expires,
	}

	return c.Status(fiber.StatusOK).
		JSON(formatter.NewSuccessResponse(formatter.Success, "Login successfully", resp))
}

// @Tags         Auth
// @Summary      Logout
// @Description  Invalidate the refresh token. Requires valid refresh token in request body.
// @Accept       json
// @Produce      json
// @Param        request  body  model.LogoutRequest  true  "Request body (refresh_token)"
// @Router       /auth/logout [post]
// @Success      200  {object}  model.SuccessMessageAPIResponse
// @Failure      400  {object}  model.ErrorInvalidRequest  "Invalid request body or validation failed"
// @Failure      401  {object}  model.ErrorUnauthorized  "Invalid or expired refresh token"
// @Failure      404  {object}  model.ErrorNotFound  "Token not found"
func (a *AuthHandlerImpl) Logout(c *fiber.Ctx) error {
	req := new(model.LogoutRequest)

	if err := c.BodyParser(req); err != nil {
		golog.Error("Error parsing request body", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := a.AuthService.Logout(c, req); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(formatter.NewSuccessResponse(formatter.Success, "Logout successfully", nil))
}

// @Tags         Auth
// @Summary      Refresh auth tokens
// @Description  Exchange a valid refresh token for a new access token.
// @Accept       json
// @Produce      json
// @Param        request  body  model.RefreshTokenRequest  true  "Request body (refresh_token)"
// @Router       /auth/refresh-tokens [post]
// @Success      200  {object}  model.RefreshTokenResponse
// @Failure      400  {object}  model.ErrorInvalidRequest  "Invalid request body or validation failed"
// @Failure      401  {object}  model.ErrorUnauthorized  "Invalid or expired refresh token"
// @Failure      404  {object}  model.ErrorNotFound  "Token not found"
func (a *AuthHandlerImpl) RefreshTokens(c *fiber.Ctx) error {
	req := new(model.RefreshTokenRequest)

	if err := c.BodyParser(req); err != nil {
		golog.Error("Error parsing request body", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	accessToken, err := a.AuthService.RefreshAuth(c, req)
	if err != nil {
		return err
	}

	resp := &model.RefreshTokenResponse{
		AccessToken: accessToken.Token,
		ExpiresAt:   accessToken.Expires,
	}

	return c.Status(fiber.StatusOK).
		JSON(formatter.NewSuccessResponse(formatter.Success, "Refresh tokens successfully", resp))
}

// @Tags         Auth
// @Summary      Forgot password
// @Description  Send a reset password email. Requires email and current password for verification.
// @Accept       json
// @Produce      json
// @Param        request  body  model.ForgotPasswordRequest  true  "Request body (email, password for verification)"
// @Router       /auth/forgot-password [post]
// @Success      200  {object}  model.SuccessMessageAPIResponse
// @Failure      400  {object}  model.ErrorInvalidRequest  "Invalid request body or validation failed"
// @Failure      401  {object}  model.ErrorFailedLogin  "Invalid email or password"
// @Failure      404  {object}  model.ErrorNotFound  "User not found"
func (a *AuthHandlerImpl) ForgotPassword(c *fiber.Ctx) error {
	req := new(model.ForgotPasswordRequest)

	if err := c.BodyParser(req); err != nil {
		golog.Error("Error parsing request body", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	resetPasswordToken, err := a.TokenService.GenerateResetPasswordToken(c, req)
	if err != nil {
		return err
	}

	if err := a.EmailAdapter.SendResetPasswordEmail(req.Email, resetPasswordToken.Token); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(formatter.NewSuccessResponse(formatter.Success, "Reset password email sent successfully", nil))
}

// @Tags         Auth
// @Summary      Reset password
// @Description  Reset password using the token from the forgot-password email.
// @Accept       json
// @Produce      json
// @Param        request  body  model.ResetPasswordRequest  true  "Request body (token from email, new password)"
// @Router       /auth/reset-password [post]
// @Success      200  {object}  model.SuccessMessageAPIResponse
// @Failure      400  {object}  model.ErrorInvalidRequest  "Invalid request body or validation failed"
// @Failure      401  {object}  model.ErrorFailedResetPassword  "Invalid or expired reset token"
// @Failure      404  {object}  model.ErrorNotFound  "User not found"
func (a *AuthHandlerImpl) ResetPassword(c *fiber.Ctx) error {
	req := new(model.ResetPasswordRequest)

	if err := c.BodyParser(req); err != nil {
		golog.Error("Error parsing request body", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := a.AuthService.ResetPassword(c, req); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(formatter.NewSuccessResponse(formatter.Success, "Update password successfully", nil))
}

// @Tags         Auth
// @Summary      Send verification email
// @Description  Send a verification email to the authenticated user. Requires Bearer token.
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Router       /auth/send-verification-email [post]
// @Success      200  {object}  model.SuccessMessageAPIResponse
// @Failure      401  {object}  model.ErrorUnauthorized  "Invalid or missing access token"
func (a *AuthHandlerImpl) SendVerificationEmail(c *fiber.Ctx) error {
	userID := c.Locals("user").(*domain.User).ID.String()

	verifyEmailToken, err := a.TokenService.GenerateVerifyEmailToken(c, userID)
	if err != nil {
		return err
	}

	if err := a.EmailAdapter.SendVerificationEmail(userID, verifyEmailToken.Token); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(formatter.NewSuccessResponse(formatter.Success, "Verification email sent successfully", nil))
}

// @Tags         Auth
// @Summary      Verify email
// @Description  Verify email address using the token from the verification email.
// @Accept       json
// @Produce      json
// @Param        request  body  model.VerifyEmailRequest  true  "Request body (token from verification email)"
// @Router       /auth/verify-email [post]
// @Success      200  {object}  model.SuccessMessageAPIResponse
// @Failure      400  {object}  model.ErrorInvalidRequest  "Invalid request body or validation failed"
// @Failure      401  {object}  model.ErrorFailedVerifyEmail  "Invalid or expired verification token"
// @Failure      404  {object}  model.ErrorNotFound  "User not found"
func (a *AuthHandlerImpl) VerifyEmail(c *fiber.Ctx) error {
	req := new(model.VerifyEmailRequest)

	if err := c.BodyParser(req); err != nil {
		golog.Error("Error parsing request body", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := a.AuthService.VerifyEmail(c, req); err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(formatter.NewSuccessResponse(formatter.Success, "Verify email successfully", nil))
}

// @Tags         Auth
// @Summary      Login with google
// @Description  Initiates the Google OAuth2 login flow. Redirects to Google consent page. Use in browser.
// @Produce      json
// @Router       /auth/google [get]
// @Success      303  "Redirects to Google OAuth consent page"
func (a *AuthHandlerImpl) GoogleLogin(c *fiber.Ctx) error {
	// Generate a random state
	state := uuid.New().String()

	c.Cookie(&fiber.Cookie{
		Name:   "oauth_state",
		Value:  state,
		MaxAge: 30,
	})

	url := a.GoogleAdapter.AuthCodeURL(state)

	return c.Status(fiber.StatusSeeOther).Redirect(url)
}

// @Tags         Auth
// @Summary      Google OAuth2 callback
// @Description  OAuth2 callback. Exchanges code for tokens and creates/updates user. Returns auth tokens.
// @Produce      json
// @Param        code   query  string  true   "Authorization code from Google"
// @Param        state  query  string  true   "State parameter for CSRF protection"
// @Router       /auth/google-callback [get]
// @Success      200  {object}  model.LoginResponse
// @Failure      401  {object}  model.ErrorUnauthorized  "Invalid state or authorization code"
func (a *AuthHandlerImpl) GoogleCallback(c *fiber.Ctx) error {
	state := c.Query("state")
	storedState := c.Cookies("oauth_state")

	if state != storedState {
		return fiber.NewError(fiber.StatusUnauthorized, "States is not match!")
	}

	code := c.Query("code")
	oauthToken, err := a.GoogleAdapter.Exchange(c.Context(), code)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		c.Context(), http.MethodGet,
		"https://www.googleapis.com/oauth2/v2/userinfo?access_token="+oauthToken.AccessToken,
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	userData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	googleUser := new(model.CreateGoogleUserRequest)
	if errJSON := json.Unmarshal(userData, googleUser); errJSON != nil {
		return errJSON
	}

	user, err := a.UserService.CreateGoogleUser(c, googleUser)
	if err != nil {
		return err
	}

	accessToken, refreshToken, err := a.TokenService.GenerateAuthTokens(c, user.ID.String())
	if err != nil {
		return err
	}

	callbackResp := &model.LoginResponse{
		AccessToken:           accessToken.Token,
		AccessTokenExpiresAt:  accessToken.Expires,
		RefreshToken:          refreshToken.Token,
		RefreshTokenExpiresAt: refreshToken.Expires,
	}

	return c.Status(fiber.StatusOK).
		JSON(formatter.NewSuccessResponse(formatter.Success, "Login successfully", callbackResp))

	// TODO: replace this url with the link to the oauth google success page of your front-end app
	// googleLoginURL := fmt.Sprintf("http://link-to-app/google/success?access_token=%s&refresh_token=%s",
	// 	tokens.Access.Token, tokens.Refresh.Token)

	// return c.Status(fiber.StatusSeeOther).Redirect(googleLoginURL)
}
