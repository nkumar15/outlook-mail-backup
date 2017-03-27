package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	HOST          = "https://login.microsoftonline.com/"
	TENANT        = "consumers"
	AUTH_URI      = "/oauth2/v2.0/authorize?"
	RESPONSE_TYPE = "code"
	RESPONSE_MODE = "query"
	STATE         = "12345"

	TOKEN_URI     = "/oauth2/v2.0/token"
	CLIENT_ID     = "e070a02d-b02f-4698-9db7-b75f4b0f30d6"
	SCOPE         = "Mail.Read"
	REDIRECT_URI  = "http://localhost:5000/auth/callback/outlook"
	GRANT_TYPE    = "authorization_code"
	CLIENT_SECRET = "HRFahPr05inQgK8vJSYMoxQ"
)

func getAuthCode() (string, error) {

	request := HOST + TENANT + AUTH_URI +
		"client_id=" + CLIENT_ID +
		`&response_type=` + RESPONSE_TYPE +
		`&redirect_uri=` + REDIRECT_URI +
		`&response_mode=` + RESPONSE_MODE +
		`&state=` + STATE +
		`&scope=` + SCOPE

	resp, err := http.Get(request)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		log.Println("Response error")
		log.Println(string(resp.StatusCode) + ":" + resp.Status)
		return "", err
	}

	defer resp.Body.Close()

	fmt.Println("Please enter the following URL in your browser and copy the code from URL and paste here")
	fmt.Println(request)

	fmt.Println("Enter code")
	var code string
	_, err = fmt.Scan(&code)

	if err != nil {
		log.Fatal("Error reading authorization code from console")
		return "", err
	}

	return code, nil
}

func getAccessToken(authCode string) (string, error) {
	hostUrl := HOST + TENANT + TOKEN_URI

	body := url.Values{}
	body.Add("client_id", CLIENT_ID)
	body.Add("code", authCode)
	body.Add("redirect_uri", REDIRECT_URI)
	body.Add("grant_type", GRANT_TYPE)
	body.Add("client_secret", CLIENT_SECRET)

	buf := bytes.NewBufferString(body.Encode())

	client := &http.Client{}
	req, _ := http.NewRequest("POST", hostUrl, buf)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New(resp.Status)
	}

	defer resp.Body.Close()
	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	respBody := string(respBodyBytes[:])
	return respBody, nil

}

func main() {

	authCode, err := getAuthCode()
	if err != nil {
		log.Println("Trouble getting Authorization code:", err)
		return
	}
	fmt.Println("AuthCode", authCode)
	accessToken, err := getAccessToken(authCode)
	if err != nil {
		log.Println("Trouble getting Access token", err)
		return
	}

	fmt.Println(accessToken)

}
