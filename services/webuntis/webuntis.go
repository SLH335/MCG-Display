package webuntis

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fastjson"
)

// constants specific to MCG
const baseUrl string = "https://herakles.webuntis.com/"
const schoolName string = "Marie-Curie-Gym"
const schoolNameBase64 string = "_bWFyaWUtY3VyaWUtZ3lt"
const appId string = "MCG-Display"

type Session struct {
	ClassId      int
	PersonId     int
	PersonType   int
	SessionId    string
	SessionToken string
}

// structs used for json encoding of auth request bodies
type authRequestBody struct {
	Id      string                `json:"id"`
	Method  string                `json:"method"`
	Params  authRequestBodyParams `json:"params"`
	JsonRpc string                `json:"jsonrpc"`
}

type authRequestBodyParams struct {
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	Client   string `json:"client,omitempty"`
}

func buildAuthRequestBody(method, username, password string) io.Reader {
	client := ""
	if username != "" || password != "" {
		client = appId
	}

	body := authRequestBody{
		Id:     appId,
		Method: method,
		Params: authRequestBodyParams{
			User:     username,
			Password: password,
			Client:   client,
		},
		JsonRpc: "2.0",
	}

	jsonBody, _ := json.Marshal(body)

	return bytes.NewReader(jsonBody)
}

func (session *Session) buildCookies() string {
	cookies := []string{
		"JSESSIONID=" + session.SessionId,
		"schoolname=" + schoolName,
	}
	return strings.Join(cookies, "; ")
}

// get new session token for requests that require authorization
func (session *Session) getSessionToken() (token string, err error) {
	token, err = session.Request(http.MethodGet, "WebUntis/api/token/new", nil, nil, false)

	session.SessionToken = token
	return token, err
}

func (session *Session) GetSessionToken() (token string, err error) {
	token, err = session.Request(http.MethodGet, "WebUntis/api/token/new", nil, nil, false)

	session.SessionToken = token
	return token, err
}

// check if session token is valid
func (session *Session) isSessionTokenValid() bool {
	if session.SessionToken == "" {
		return false
	}

	tokenParts := strings.Split(session.SessionToken, ".")
	if len(tokenParts) != 3 {
		return false
	}

	data, err := base64.RawStdEncoding.DecodeString(tokenParts[1])
	if err != nil || len(data) == 0 {
		return false
	}

	expires := time.Unix(int64(fastjson.GetInt(data, "exp")), 0)

	return expires.After(time.Now())
}

// generic request to the WebUntis API
func (session *Session) Request(method, url string, queryParams url.Values, jsonBody []byte, auth bool) (result string, err error) {
	req, err := http.NewRequest(method, baseUrl+url+"?"+queryParams.Encode(), bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Add("Cookie", session.buildCookies())
	if auth {
		if !session.isSessionTokenValid() {
			session.getSessionToken()
		}
		req.Header.Add("Authorization", "Bearer "+session.SessionToken)
	}
	if len(jsonBody) > 0 {
		req.Header.Add("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(resBody), nil
}

// create a new WebUntis session
func Login(username, password string) (session Session, err error) {
	url := baseUrl + "WebUntis/jsonrpc.do?school=" + schoolName
	reqBody := buildAuthRequestBody("authenticate", username, password)

	res, err := http.Post(url, "application/json", reqBody)
	if err != nil {
		return session, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return session, err
	}

	var parser fastjson.Parser
	jsonData, err := parser.Parse(string(resBody))
	if err != nil {
		return session, err
	}

	session = Session{
		ClassId:    jsonData.GetInt("result", "klasseId"),
		PersonId:   jsonData.GetInt("result", "personId"),
		PersonType: jsonData.GetInt("result", "personType"),
		SessionId:  string(jsonData.GetStringBytes("result", "sessionId")),
	}

	return session, err
}

func (session *Session) Logout() error {
	url := baseUrl + "WebUntis/jsonrpc.do?school=" + schoolName
	reqBody := buildAuthRequestBody("logout", "", "")

	_, err := http.Post(url, "application/json", reqBody)
	if err != nil {
		return err
	}

	session = &Session{}
	return nil
}

func convertDateToUntis(date time.Time) string {
	return date.Format("2006-01-02")
}

func (session *Session) GetExams(start, end time.Time, withDeleted bool) (exams []Exam, err error) {
	path := "WebUntis/api/rest/view/v1/exams"
	queryParams := url.Values{
		"start":       {convertDateToUntis(start)},
		"end":         {convertDateToUntis(end)},
		"withDeleted": {strconv.FormatBool(withDeleted)},
	}

	res, err := session.Request(http.MethodGet, path, queryParams, nil, true)
	if err != nil {
		return exams, err
	}

	var jsonData struct {
		Exams       []Exam `json:"exams"`
		WithDeleted bool   `json:"withDeleted"`
	}
	err = json.Unmarshal([]byte(res), &jsonData)
	if err != nil {
		return exams, err
	}
	exams = jsonData.Exams

	return exams, err
}
