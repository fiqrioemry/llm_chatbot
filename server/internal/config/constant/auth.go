package constant


const (
	RegisterSuccess              = "Registration successful. Please check your email to verify your account."
	LoginSuccess                 = "Login successful"
	LogoutSuccess                = "Logout successful"
	LogoutAllSuccess             = "All sessions revoked"
	EmailVerificationSuccess     = "Email verified successfully"
	ResendVerificationSuccess    = "Verification email resent"
	ChangePasswordSuccess        = "Password changed successfully"
	ForgotPasswordSuccess        = "If the email is registered, a reset link has been sent"
	ResetPasswordSuccess         = "Password reset successfully"
	GetSessionsSuccess           = "Sessions retrieved successfully"
	DeleteSessionSuccess         = "Session deleted successfully"
	GetMeSuccess                 = "User retrieved successfully"
	RefreshTokenSuccess          = "Token refreshed successfully"
)
 
const (
	ErrUnauthorized            = "Unauthorized"
	ErrForbidden               = "Forbidden"
	ErrInvalidToken            = "Invalid or expired token"
	ErrTokenRequired           = "Token is required"
	ErrSessionRevoked          = "Session has been revoked"
	ErrAccountBanned           = "Your account has been banned"
	ErrEmailExists             = "Email is already registered"
	ErrEmailNotVerified        = "Please verify your email before logging in"
	ErrEmailAlreadyVerified    = "Email is already verified"
	ErrInvalidCredentials      = "Invalid email or password"
	ErrUserNotFound            = "User not found"
	ErrSessionNotFound         = "Session not found"
	ErrUserCreationFailed      = "Failed to create user"
	ErrTooManyRequests         = "Too many requests, please try again later"
	ErrBadRequest              = "Bad request"
	ErrInternalServerError     = "Internal server error"
	ErrRouteNotFound           = "Route not found"
	ErrOAuthStateMismatch      = "OAuth state mismatch, possible CSRF attack"
	ErrOAuthFailed             = "OAuth authentication failed"
	ErrOAuthEmailNotProvided   = "OAuth provider did not return an email"
	ErrRefreshTokenInvalid     = "Refresh token is invalid or expired"
	ErrRefreshTokenRequired    = "Refresh token is required"
	ErrCannotFollowSelf        = "You cannot follow yourself"
	ErrAlreadyFollowing        = "You are already following this user"
	ErrFollowRequestSent       = "Follow request already sent"
)


const (
	CodeUnauthorized          = "UNAUTHORIZED"
	CodeForbidden             = "FORBIDDEN"
	CodeInvalidToken          = "INVALID_TOKEN"
	CodeTokenRequired         = "TOKEN_REQUIRED"
	CodeSessionRevoked        = "SESSION_REVOKED"
	CodeAccountBanned         = "ACCOUNT_BANNED"
	CodeEmailExists           = "EMAIL_EXISTS"
	CodeEmailNotVerified      = "EMAIL_NOT_VERIFIED"
	CodeEmailAlreadyVerified  = "EMAIL_ALREADY_VERIFIED"
	CodeInvalidCredentials    = "INVALID_CREDENTIALS"
	CodeUserNotFound          = "USER_NOT_FOUND"
	CodeSessionNotFound       = "SESSION_NOT_FOUND"
	CodeUserCreationFailed    = "USER_CREATION_FAILED"
	CodeTooManyRequests       = "TOO_MANY_REQUESTS"
	CodeBadRequest            = "BAD_REQUEST"
	CodeInternalServerError   = "INTERNAL_SERVER_ERROR"
	CodeRouteNotFound         = "ROUTE_NOT_FOUND"
	CodeOAuthStateMismatch    = "OAUTH_STATE_MISMATCH"
	CodeOAuthFailed           = "OAUTH_FAILED"
	CodeOAuthEmailNotProvided = "OAUTH_EMAIL_NOT_PROVIDED"
	CodeRefreshTokenInvalid   = "REFRESH_TOKEN_INVALID"
	CodeRefreshTokenRequired  = "REFRESH_TOKEN_REQUIRED"
)

 
type RateLimitConfig struct {
	Limit     int
	WindowSec int
}

const (
	ActionLogin			= "login"
	ActionRegister			= "register"
)

var (
	LimitLogin = RateLimitConfig{
		Limit:     10,
		WindowSec: 60,
	}
	LimitRegister = RateLimitConfig{
		Limit:     5,
		WindowSec: 60,
	}
	LimitSessions = RateLimitConfig{
		Limit:     20,
		WindowSec: 60,
	}
	LimitPasswordChange = RateLimitConfig{
		Limit:     5,
		WindowSec: 60,
	}
	LimitPasswordReset = RateLimitConfig{
		Limit:     5,
		WindowSec: 60,
	}
	LimitPasswordForgot = RateLimitConfig{
		Limit:     5,
		WindowSec: 60,
	}
	LimitResendVerification = RateLimitConfig{
		Limit:     3,
		WindowSec: 60,
	}
	LimitVerifyEmail = RateLimitConfig{
		Limit:     10,
		WindowSec: 60,
	}
	LimitGetMe = RateLimitConfig{
		Limit:     30,
		WindowSec: 60,
	}
	LimitRefreshToken = RateLimitConfig{
		Limit:     20,
		WindowSec: 60,
	}
)