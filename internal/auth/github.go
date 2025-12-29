package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	ghDeviceCodeURL  = "https://github.com/login/device/code"
	ghAccessTokenURL = "https://github.com/login/oauth/access_token"
	ghUserURL        = "https://api.github.com/user"
)

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error,omitempty"`
}

type GithubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func RequestDeviceCode(clientID string) (*DeviceCodeResponse, error) {
	body := map[string]string{
		"client_id": clientID,
		"scope":     "read:user user:email",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", ghDeviceCodeURL, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DeviceCodeResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}

func PollAccessToken(clientID, clientSecret, deviceCode string) (*AccessTokenResponse, error) {
	body := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"device_code":   deviceCode,
		"grant_type":    "urn:ietf:params:oauth:grant-type:device_code",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", ghAccessTokenURL, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result AccessTokenResponse
	json.NewDecoder(resp.Body).Decode(&result)
	return &result, nil
}

func GetGithubUser(accessToken string) (*GithubUser, error) {
	req, _ := http.NewRequest("GET", ghUserURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %d", resp.StatusCode)
	}

	var user GithubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode github user: %w", err)
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("invalid github user: ID is 0")
	}

	return &user, nil
}
