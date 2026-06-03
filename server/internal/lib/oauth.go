package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"server/internal/config"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

 
type OAuthUserInfo struct {
	ProviderAccountID string
	Email             string
	Name              string
	Avatar            string
}

 
type OAuthTokens struct {
	AccessToken  string
	RefreshToken *string
	ExpiresAt    *time.Time
}
 
type OAuthProvider interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*OAuthTokens, error)
	GetUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error)
}

 
type OAuthRegistry struct {
	providers map[string]OAuthProvider
}

func NewOAuthRegistry(cfg *config.Config) *OAuthRegistry {
	r := &OAuthRegistry{providers: make(map[string]OAuthProvider)}

	if cfg.OAuth.GoogleClientID != "" {
		r.providers["google"] = newGoogleProvider(cfg)
	}
	if cfg.OAuth.GithubClientID != "" {
		r.providers["github"] = newGithubProvider(cfg)
	}

	return r
}

func (r *OAuthRegistry) Get(provider string) (OAuthProvider, error) {
	p, ok := r.providers[provider]
	if !ok {
		return nil, fmt.Errorf("oauth provider '%s' is not configured", provider)
	}
	return p, nil
}

 
type googleProvider struct{ cfg *oauth2.Config }

func newGoogleProvider(cfg *config.Config) *googleProvider {
	return &googleProvider{cfg: &oauth2.Config{
		ClientID:     cfg.OAuth.GoogleClientID,
		ClientSecret: cfg.OAuth.GoogleClientSecret,
		RedirectURL:  cfg.OAuth.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}}
}

func (g *googleProvider) GetAuthURL(state string) string {
	return g.cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (g *googleProvider) ExchangeCode(ctx context.Context, code string) (*OAuthTokens, error) {
	t, err := g.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	tokens := &OAuthTokens{AccessToken: t.AccessToken}
	if t.RefreshToken != "" {
		tokens.RefreshToken = &t.RefreshToken
	}
	if !t.Expiry.IsZero() {
		tokens.ExpiresAt = &t.Expiry
	}
	return tokens, nil
}

func (g *googleProvider) GetUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var data struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return &OAuthUserInfo{
		ProviderAccountID: data.ID,
		Email:             data.Email,
		Name:              data.Name,
		Avatar:            data.Picture,
	}, nil
}
 
type githubProvider struct{ cfg *oauth2.Config }

func newGithubProvider(cfg *config.Config) *githubProvider {
	return &githubProvider{cfg: &oauth2.Config{
		ClientID:     cfg.OAuth.GithubClientID,
		ClientSecret: cfg.OAuth.GithubClientSecret,
		RedirectURL:  cfg.OAuth.GithubRedirectURL,
		Scopes:       []string{"user:email", "read:user"},
		Endpoint:     github.Endpoint,
	}}
}

func (g *githubProvider) GetAuthURL(state string) string {
	return g.cfg.AuthCodeURL(state)
}

func (g *githubProvider) ExchangeCode(ctx context.Context, code string) (*OAuthTokens, error) {
	t, err := g.cfg.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	return &OAuthTokens{AccessToken: t.AccessToken}, nil
}

func (g *githubProvider) GetUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	doReq := func(url string) ([]byte, error) {
		req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/vnd.github+json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return io.ReadAll(resp.Body)
	}

	body, err := doReq("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	var profile struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Email     string `json:"email"`
	}
	json.Unmarshal(body, &profile)

	// Ambil primary email jika tidak tersedia di profile
	email := profile.Email
	if email == "" {
		emailBody, err := doReq("https://api.github.com/user/emails")
		if err == nil {
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			json.Unmarshal(emailBody, &emails)
			for _, e := range emails {
				if e.Primary {
					email = e.Email
					break
				}
			}
		}
	}

	name := profile.Name
	if name == "" {
		name = profile.Login
	}

	return &OAuthUserInfo{
		ProviderAccountID: fmt.Sprintf("%d", profile.ID),
		Email:             email,
		Name:              name,
		Avatar:            profile.AvatarURL,
	}, nil
}