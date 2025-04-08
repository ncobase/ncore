package oauth

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/google/go-querystring/query"
)

// Picture defines the structure for a Facebook profile picture.
type Picture struct {
	Data struct {
		Height       int    `json:"height"`
		IsSilhouette bool   `json:"is_silhouette"`
		URL          string `json:"url"`
		Width        int    `json:"width"`
	} `json:"data"`
}

// FacebookProfile represents a Facebook user profile.
type FacebookProfile struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Picture Picture `json:"picture"`
}

// FacebookToken represents a Facebook OAuth token.
type FacebookToken struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	Raw          any       `json:"raw"`
}

// FacebookTokenJSON is used for unmarshalling the JSON response from Facebook.
type FacebookTokenJSON struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
}

// expiry calculates the expiry time for the token.
func (e *FacebookTokenJSON) expiry() time.Time {
	if v := e.ExpiresIn; v != 0 {
		return time.Now().Add(time.Duration(v) * time.Second)
	}
	return time.Time{}
}

// GetFacebookAccessToken retrieves the Facebook access token using the provided authorization code.
func GetFacebookAccessToken(redirectURL, code string) (string, error) {
	clientID := os.Getenv("FACEBOOK_CLIENT_ID")
	clientSecret := os.Getenv("FACEBOOK_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", errors.New("Facebook client ID or secret not set in environment variables")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	type FacebookOAuthParams struct {
		Code         string `url:"code" json:"code"`
		ClientID     string `url:"client_id" json:"client_id"`
		ClientSecret string `url:"client_secret" json:"client_secret"`
		RedirectURI  string `url:"redirect_uri" json:"redirect_uri"`
	}

	p := FacebookOAuthParams{
		Code:         code,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURL + "/v1/oauth/callback/facebook",
	}

	queryString, err := query.Values(p)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("GET", "https://graph.facebook.com/v4.0/oauth/access_token?"+queryString.Encode(), nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}

	token, err := parseFacebookToken(body, resp.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	if token.AccessToken == "" {
		return "", errors.New("oauth: server response missing access_token")
	}
	return token.AccessToken, nil
}

// parseFacebookToken parses the token from the response body and content type.
func parseFacebookToken(body []byte, contentType string) (*FacebookToken, error) {
	var token *FacebookToken
	content, _, _ := mime.ParseMediaType(contentType)
	switch content {
	case "application/x-www-form-urlencoded", "text/plain":
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return nil, err
		}

		token = &FacebookToken{
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
	default:
		var tokenJSON FacebookTokenJSON
		if err := json.Unmarshal(body, &tokenJSON); err != nil {
			return nil, err
		}

		token = &FacebookToken{
			AccessToken:  tokenJSON.AccessToken,
			TokenType:    tokenJSON.TokenType,
			RefreshToken: tokenJSON.RefreshToken,
			Expiry:       tokenJSON.expiry(),
			Raw:          make(map[string]any),
		}
		if err := json.Unmarshal(body, &token.Raw); err != nil {
			return nil, err
		}
	}
	return token, nil
}

// GetFacebookProfile retrieves the user's Facebook profile using the provided access token.
func GetFacebookProfile(token string) (*Profile, error) {
	req, err := http.NewRequest("GET", "https://graph.facebook.com/v4.0/me?fields=id,name,email,picture", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result FacebookProfile
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	profile := Profile{
		ID:        result.ID,
		Name:      result.Name,
		Email:     result.Email,
		Thumbnail: result.Picture.Data.URL,
	}

	return &profile, nil
}
