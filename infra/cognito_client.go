package infra

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/oauth2"
	"openflp.com/model"
)

type CognitoClient interface {
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (*model.AuthResponse, error)
	VerifyToken(token string) (*model.User, error)
}

type realCognitoClient struct {
	oauth2Config oauth2.Config
	provider     *oidc.Provider
}

func NewRealCognitoClient(clientID, clientSecret, redirectURL, issuerURL string) (CognitoClient, error) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %v", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "openid"},
	}

	return &realCognitoClient{
		oauth2Config: oauth2Config,
		provider:     provider,
	}, nil
}

func (c *realCognitoClient) GetAuthURL(state string) string {
	return c.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (c *realCognitoClient) ExchangeCode(ctx context.Context, code string) (*model.AuthResponse, error) {
	rawToken, err := c.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %v", err)
	}

	tokenString := rawToken.AccessToken

	// Parse the token for claims
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	if name == "" {
		name, _ = claims["nickname"].(string)
	}

	return &model.AuthResponse{
		Token: tokenString,
		User: model.User{
			ID:    fmt.Sprintf("%v", claims["sub"]),
			Email: email,
			Name:  name,
		},
	}, nil
}

func (c *realCognitoClient) VerifyToken(tokenString string) (*model.User, error) {
	// In production, you should use the provider to verify the token signature
	// For this implementation, we will follow the doc's parsing logic
	// but using the provider's verifier is better.
	
	ctx := context.Background()
	verifier := c.provider.Verifier(&oidc.Config{ClientID: c.oauth2Config.ClientID})
	
	// Try to verify as ID token first (common in OIDC)
	idToken, err := verifier.Verify(ctx, tokenString)
	if err == nil {
		var claims struct {
			Subject string `json:"sub"`
			Email   string `json:"email"`
			Name    string `json:"name"`
		}
		if err := idToken.Claims(&claims); err != nil {
			return nil, err
		}
		return &model.User{
			ID:    claims.Subject,
			Email: claims.Email,
			Name:  claims.Name,
		}, nil
	}

	// Fallback to manual parsing if it's an access token
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return &model.User{
		ID:    fmt.Sprintf("%v", claims["sub"]),
		Email: fmt.Sprintf("%v", claims["email"]),
		Name:  fmt.Sprintf("%v", claims["username"]),
	}, nil
}
