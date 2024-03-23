package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo"
)

type OAuthAccessToken struct {
	AccessToken   string `json:"access_token,omitempty"`
	WorkspaceName string `json:"workspace_name,omitempty"`
	WorkspaceIcon string `json:"workspace_icon,omitempty"`
	BotID         string `json:"bot_id,omitempty"`
}

var (
	OAUTH_CLIENT_SECRET = "OAUTH_CLIENT_SECRET"
	OAUTH_CLIENT_ID     = "OAUTH_CLIENT_ID"
	REDIRECT_URL        = "REDIRECT_URL"
)

func Handler(c echo.Context) (string, string, error) {
	oauthClientSecret := os.Getenv(OAUTH_CLIENT_SECRET)
	oauthClientId := os.Getenv(OAUTH_CLIENT_ID)
	redirectUrl := os.Getenv(REDIRECT_URL)

	client := http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       time.Second * 5,
	}

	code := c.QueryParam("code")
	log.Println("got request with code ", code)

	stateToken := c.QueryParam("state")
	b, err := json.Marshal(&struct {
		GrantType   string `json:"grant_type,omitempty"`
		Code        string `json:"code,omitempty"`
		RedirectURI string `json:"redirect_uri,omitempty"`
	}{
		GrantType:   "authorization_code",
		Code:        code,
		RedirectURI: redirectUrl,
	})
	if err != nil {
		log.Fatal(err)
		return "", "", err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://api.notion.com/v1/oauth/token", bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
		return "", "", err
	}
	req.SetBasicAuth(oauthClientId, oauthClientSecret)
	req.Header.Add("Content-Type", "application/json")

	log.Println("got request:", req)

	rsp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
		return "", "", err
	}

	defer rsp.Body.Close()

	var body OAuthAccessToken
	if err = json.NewDecoder(rsp.Body).Decode(&body); err != nil {
		fmt.Println(err)
		return "", "", err
	}

	return body.AccessToken, stateToken, nil
}
