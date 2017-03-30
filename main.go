package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	AUTH_HOST     = "https://login.microsoftonline.com/"
	TENANT        = "consumers"
	AUTH_URI      = "/oauth2/v2.0/authorize?"
	RESPONSE_TYPE = "code"
	RESPONSE_MODE = "query"
	STATE         = "12345"

	TOKEN_URI     = "/oauth2/v2.0/token"
	CLIENT_ID     = "e070a02d-b02f-4698-9db7-b75f4b0f30d6"
	SCOPE         = "User.Read"
	REDIRECT_URI  = "http://localhost:5000/auth/callback/outlook"
	GRANT_TYPE    = "authorization_code"
	CLIENT_SECRET = "HRFahPr05inQgK8vJSYMoxQ"

	GRAPH_HOST    = "https://graph.microsoft.com/"
	GRAPH_VERSION = "v1.0/"
)

type GraphToken struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    uint   `json:"expires_in"`
	ExtExpiresIn uint   `json:"ext_expires_in"`
	AccessToken  string `json:"access_token"`
}

type User struct {
	Name          string `json:"givenName"`
	SurName       string `json:"surname"`
	DisplayName   string `json:"displayName"`
	Id            string `json:"id"`
	PrincipleName string `json:"userPrincipalName"`
}

type EmailAddress struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type SenderItem struct {
	Email EmailAddress `json:"emailAddress"`
}

type FromItem struct {
	Email EmailAddress `json:"emailAddress"`
}

type RecipientItems struct {
	Email EmailAddress `json:"emailAddress"`
}

type EmailBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type Message struct {
	Id           string    `json:"id"`
	CreateDate   string    `json:"createdDateTime"`
	SentDate     string    `json:"sentDateTime"`
	RecievedDate string    `json:"receivedDateTime"`
	Subject      string    `json:"subject"`
	Preview      string    `json:"bodyPreview"`
	IsRead       bool      `json:"isRead"`
	IsDraft      bool      `json:"isDraft"`
	Body         EmailBody `json:"body"`
	Attachments  bool      `json:"hasAttachments"`
}

type MessageList struct {
	Context    string           `json:"@odata.context"`
	NextLink   string           `json:"@odata.nextLink"`
	Message    []Message        `json:"value"`
	Sender     SenderItem       `json:"sender"`
	From       FromItem         `json:"from"`
	Recipients []RecipientItems `json:"toRecipients"`
}

func getAuthCode() (string, error) {

	request := AUTH_HOST + TENANT + AUTH_URI +
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

func toGraphToken(body []byte) (*GraphToken, error) {

	var token = new(GraphToken)
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	return token, nil
}

func getAccessToken(authCode string) (*GraphToken, error) {

	hostUrl := AUTH_HOST + TENANT + TOKEN_URI

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
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var token *GraphToken
	token, err = toGraphToken(bodyBytes)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func printAccesstoken(token *GraphToken) {
	fmt.Println("Access token*****************")
	fmt.Println(token.AccessToken)
	fmt.Println("*****************************")
}

func toUser(body []byte) (*User, error) {

	var user = new(User)
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return user, nil
}

func getUserInfo(token string) (*User, error) {

	graphUrl := GRAPH_HOST + GRAPH_VERSION + "me"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", graphUrl, nil)

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var user *User
	user, err = toUser(bodyBytes)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func printUserInfo(user *User) {
	fmt.Println("User info********************")
	fmt.Println("Id: " + user.Id)
	fmt.Println("Name: " + user.Name)
	fmt.Println("Email: " + user.PrincipleName)
	fmt.Println("Display name: " + user.DisplayName)
	fmt.Println("*****************************")
}

func toMessageList(body []byte) (*MessageList, error) {

	var msgList = new(MessageList)
	if err := json.Unmarshal(body, &msgList); err != nil {
		return nil, err
	}

	return msgList, nil
}

func getMessageList(token string) (*MessageList, error) {

	graphUrl := GRAPH_HOST + GRAPH_VERSION + "me/messages"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", graphUrl, nil)

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var msgList *MessageList
	msgList, err = toMessageList(bodyBytes)
	if err != nil {
		return nil, err
	}

	return msgList, nil
}

func printMessageList(msgList *MessageList) {
	fmt.Println("Message Id list********************")
	for _, msg := range msgList.Message {
		fmt.Println(msg.Id)
	}

	fmt.Println("Next link: ", msgList.NextLink)
}

func toMessage(body []byte) (*Message, error) {

	var msg = new(Message)
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func getMessage(msgid, token string) (*Message, error) {

	graphUrl := GRAPH_HOST + GRAPH_VERSION + "me/messages/" + msgid

	client := &http.Client{}
	req, _ := http.NewRequest("GET", graphUrl, nil)

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var msg *Message
	msg, err = toMessage(bodyBytes)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func printMessage(msgList *MessageList, msg *Message) {
	fmt.Println("Message********************")
	fmt.Println("From:", msgList.Sender.Email.Name, "(", msgList.Sender.Email.Address, ")")

	fmt.Println("Recipients:")
	for _, rec := range msgList.Recipients {
		fmt.Println(rec.Email.Name, "(", rec.Email.Address, ")")
	}

	fmt.Println("Created ", msg.CreateDate)
	fmt.Println("Recieved ", msg.RecievedDate)
	fmt.Println("Subject:", msg.Subject)
	fmt.Println("Preview:", msg.Preview)
	fmt.Println("Body:")
	fmt.Println(msg.Body.Content)
	fmt.Println()
	fmt.Println("Attachments ", msg.Attachments)
}

func main() {

	authCode, err := getAuthCode()
	if err != nil {
		log.Println("Trouble getting Authorization code:", err)
		return
	}

	token, err := getAccessToken(authCode)
	if err != nil {
		log.Println("Trouble getting Access token", err)
		return
	}
	printAccesstoken(token)
	fmt.Println()

	user, err := getUserInfo(token.AccessToken)
	if err != nil {
		log.Println("Trouble getting User info", err)
		return
	}
	printUserInfo(user)
	fmt.Println()

	msgList, err := getMessageList(token.AccessToken)
	if err != nil {
		log.Println("Trouble getting message list", err)
		return
	}
	printMessageList(msgList)
	fmt.Println()

	fmt.Println("Enter msg id")
	var msgId string
	_, err = fmt.Scan(&msgId)

	if err != nil {
		log.Fatal("Error reading msgId from console")
		return
	}

	msg, err := getMessage(msgId, token.AccessToken)
	if err != nil {
		log.Println("Trouble getting message", err)
		return
	}
	printMessage(msgList, msg)
	fmt.Println()

}
