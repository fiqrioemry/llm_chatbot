package handlers

import (
	"log"
	"net/http"

	"server/internal/config"
	"server/internal/config/constant"
	"server/internal/dto"
	"server/internal/middleware"
	"server/internal/services"
	"server/pkg/response"
	"server/pkg/validator"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *services.AuthService
	cfg *config.Config
}

func NewAuthHandler(svc *services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{svc: svc, cfg: cfg}
}


func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c,
			constant.ErrBadRequest,
			constant.CodeBadRequest,
			validator.ExtractErrors(err),
		)
		return
	}
	log.Printf("[auth handler] Register request received for email: %s", req.Email)

	email, err := h.svc.Register(c, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[auth handler] Registration successful for email: %s", email)
	response.Created(c, constant.RegisterSuccess, gin.H{"email": email})
}


func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, constant.ErrTokenRequired, constant.CodeTokenRequired)
		return
	}
	log.Printf("[auth handler] Verify email request recieved for token: %s", req.Token)

	if err := h.svc.VerifyEmail(req.Token); err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[auth handler] Email verification successful for token: %s", req.Token)
	response.OK(c, constant.EmailVerificationSuccess)
}


func (h *AuthHandler) ResendVerificationEmail(c *gin.Context) {
	var req dto.ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c,
			constant.ErrBadRequest,
			constant.CodeBadRequest,
			validator.ExtractErrors(err),
		)
		return
	}

	log.Printf("[auth handler] Resend verification email request received for email: %s", req.Email)
	if err := h.svc.ResendVerificationEmail(req.Email); err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[auth handler] Resend verification email successful for email: %s", req.Email)
	response.OK(c, constant.ResendVerificationSuccess)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c,
			constant.ErrBadRequest,
			constant.CodeBadRequest,
			validator.ExtractErrors(err),
		)
		return
	}
	log.Printf("[auth handler] Login request received for email: %s", req.Email)
	result, err := h.svc.Login(c, req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[auth handler] Login successful for email: %s", req.Email)
	response.OK(c, constant.LoginSuccess, result)
}
 
func (h *AuthHandler) Logout(c *gin.Context) {
 
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		response.BadRequest(c, constant.ErrBadRequest, constant.CodeBadRequest)
		return
	}
	log.Printf("[auth handler] Logout request received for session ID: %s", sessionID)
	if err := h.svc.Logout(sessionID); err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[auth handler] Logout successful for session ID: %s", sessionID)
	response.OK(c, constant.LogoutSuccess)
}


func (h *AuthHandler) LogoutAll(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	log.Printf("[auth handler] Logout all sessions request received for user ID: %s", user.UserID)
	if err := h.svc.LogoutAll(user.UserID); err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[auth handler] Logout all sessions successful for user ID: %s", user.UserID)
	response.OK(c, constant.LogoutAllSuccess)
}

func (h *AuthHandler) GetSessions(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	log.Printf("[auth handler] Get sessions request received for user ID: %s", user.UserID)
	sessions, err := h.svc.GetSessions(user.UserID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	log.Printf("[auth handler] Get sessions successful for user ID: %s, found %d sessions", user.UserID, len(sessions))
	response.OK(c, constant.GetSessionsSuccess, sessions)
}


func (h *AuthHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	log.Printf("[auth handler] Delete session request received for session ID: %s", sessionID)
	if err := h.svc.DeleteSession(sessionID); err != nil {
		response.HandleError(c, err)
		return
	}
	log.Printf("[auth handler] Delete session successful for session ID: %s", sessionID)
	response.OK(c, constant.DeleteSessionSuccess)
}


func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c,
			constant.ErrRefreshTokenRequired,
			constant.CodeRefreshTokenRequired,
			validator.ExtractErrors(err),
		)
		return
	}
	log.Printf("[auth handler] Refresh token request received for refresh token: %s", req.RefreshToken)
	result, err := h.svc.RefreshToken(c, req.RefreshToken)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	log.Printf("[auth handler] Refresh token successful for refresh token: %s", req.RefreshToken)
	response.OK(c, constant.RefreshTokenSuccess, result)
}


func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c,
			constant.ErrBadRequest,
			constant.CodeBadRequest,
			validator.ExtractErrors(err),
		)
		return
	}
	log.Printf("[auth handler] Change password request received for user ID: %s", middleware.GetAuthUser(c).UserID)
	user := middleware.GetAuthUser(c)

	if err := h.svc.ChangePassword(user.UserID, req); err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[auth handler] Change password successful for user ID: %s", user.UserID)
	response.OK(c, constant.ChangePasswordSuccess)
}


func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c,
			constant.ErrBadRequest,
			constant.CodeBadRequest,
			validator.ExtractErrors(err),
		)
		return
	}
	log.Printf("[auth handler] Forgot password request received for email: %s", req.Email)
	if err := h.svc.ForgotPassword(req.Email); err != nil {
		response.HandleError(c, err)
		return
	}
	log.Printf("[auth handler] Forgot password successful for email: %s", req.Email)
	response.OK(c, constant.ForgotPasswordSuccess)
}


func (h *AuthHandler) ResetPassword(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.BadRequest(c, constant.ErrTokenRequired, constant.CodeTokenRequired)
		return
	}
	log.Printf("[auth handler] Reset password request received for token: %s", token)
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c,
			constant.ErrBadRequest,
			constant.CodeBadRequest,
			validator.ExtractErrors(err),
		)
		return
	}
	log.Printf("[auth handler] Reset password request validated for token: %s", token)
	if err := h.svc.ResetPassword(token, req); err != nil {
		response.HandleError(c, err)
		return
	}

	log.Printf("[auth handler] Reset password successful for token: %s", token)
	response.OK(c, constant.ResetPasswordSuccess)
}


func (h *AuthHandler) GetMe(c *gin.Context) {
	user := middleware.GetAuthUser(c)
	log.Printf("[auth handler] Get me request received for user ID: %s", user.UserID)
	result, err := h.svc.GetMe(user.UserID)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	log.Printf("[auth handler] Get me successful for user ID: %s", user.UserID)
	response.OK(c, constant.GetMeSuccess, result)
}


func (h *AuthHandler) OAuthRedirect(c *gin.Context) {
	provider := c.Param("provider")
	
	url, err := h.svc.GetOAuthURL(provider)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	c.Redirect(http.StatusFound, url)
}

func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code     := c.Query("code")
	state    := c.Query("state")
	oauthErr := c.Query("error")

	// User membatalkan OAuth di provider
	if oauthErr != "" {
		c.Redirect(http.StatusFound,
			h.cfg.App.ClientURL+"/login?error=oauth_cancelled&provider="+provider,
		)
		return
	}

	result, err := h.svc.HandleOAuthCallback(c, provider, code, state)
	if err != nil {
		c.Redirect(http.StatusFound,
			h.cfg.App.ClientURL+"/login?error=oauth_failed&provider="+provider,
		)
		return
	}

 
	c.Redirect(http.StatusFound,
		h.cfg.App.ClientURL+
			"/oauth/callback?success=true&provider="+provider+
			"&accessToken="+result.Tokens.AccessToken+
			"&refreshToken="+result.Tokens.RefreshToken,
	)
}

 