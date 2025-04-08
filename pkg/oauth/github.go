package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// GithubTokenSource defines the structure to hold the GitHub access token.
type GithubTokenSource struct {
	AccessToken string `json:"access_token"`
}

// GithubToken represents a GitHub OAuth token.
type GithubToken struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	Expiry       time.Time
	Raw          any
}

// GithubTokenJSON is used for unmarshalling the JSON response from GitHub.
type GithubTokenJSON struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
}

// GetGithubAccessToken retrieves the GitHub access token using the provided authorization code.
func GetGithubAccessToken(code string) (string, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", errors.New("GitHub client ID or secret not set in environment variables")
	}

	client := http.Client{Timeout: 10 * time.Second}
	v := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_secret": {clientSecret},
		"client_id":     {clientID},
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(v.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}

	var token *GithubToken
	content, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	switch content {
	case "application/x-www-form-urlencoded", "text/plain":
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return "", err
		}

		token = parseFormURLEncodedToken(values)
	default:
		token, err = parseJSONToken(body)
		if err != nil {
			return "", err
		}
	}

	if token.AccessToken == "" {
		return "", errors.New("oauth: server response missing access_token")
	}

	return token.AccessToken, nil
}

// parseFormURLEncodedToken parses the token from form-urlencoded response.
func parseFormURLEncodedToken(values url.Values) *GithubToken {
	token := &GithubToken{
		AccessToken:  values.Get("access_token"),
		TokenType:    values.Get("token_type"),
		RefreshToken: values.Get("refresh_token"),
		Raw:          values,
	}
	if expires := values.Get("expires_in"); expires != "" {
		if expiresIn, err := strconv.Atoi(expires); err == nil && expiresIn != 0 {
			token.Expiry = time.Now().Add(time.Duration(expiresIn) * time.Second)
		}
	}
	return token
}

// parseJSONToken parses the token from JSON response.
func parseJSONToken(body []byte) (*GithubToken, error) {
	var tokenJSON GithubTokenJSON
	if err := json.Unmarshal(body, &tokenJSON); err != nil {
		return nil, err
	}

	token := &GithubToken{
		AccessToken:  tokenJSON.AccessToken,
		TokenType:    tokenJSON.TokenType,
		RefreshToken: tokenJSON.RefreshToken,
		Expiry:       tokenJSON.expiry(),
		Raw:          make(map[string]any),
	}
	_ = json.Unmarshal(body, &token.Raw)
	return token, nil
}

// expiry calculates the expiry time for the token.
func (e *GithubTokenJSON) expiry() time.Time {
	if e.ExpiresIn != 0 {
		return time.Now().Add(time.Duration(e.ExpiresIn) * time.Second)
	}
	return time.Time{}
}

// Token returns the OAuth2 token for the GitHub token source.
func (t *GithubTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{AccessToken: t.AccessToken}, nil
}

// GetGithubProfile retrieves the user's GitHub profile using the provided access token.
func GetGithubProfile(accessToken string) (*Profile, error) {
	ctx := context.Background()
	oauthClient := oauth2.NewClient(ctx, &GithubTokenSource{AccessToken: accessToken})
	client := github.NewClient(oauthClient)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	profile := Profile{
		ID:        strconv.FormatInt(*user.ID, 10),
		Name:      getString(user.Name),
		Email:     getString(user.Email),
		Thumbnail: getString(user.AvatarURL),
	}

	return &profile, nil
}

// getString safely dereferences a string pointer.
func getString(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}
