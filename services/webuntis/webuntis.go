package webuntis

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/valyala/fastjson"
)

// constants specific to MCG
const baseUrl string = "https://herakles.webuntis.com/"
const schoolName string = "Marie-Curie-Gym"
const schoolNameBase64 string = "_bWFyaWUtY3VyaWUtZ3lt"
const appId string = "MCG-Display"

// timetable resources to access calendar; must be accessible by the user issuing the request
const calendarResourceType string = "STUDENT"
const calendarResource int = 5186

type Session struct {
	ClassId      int
	PersonId     int
	PersonType   int
	SessionId    string
	SessionToken string
}

type authRequestType int

const (
	passwordAuthRequest authRequestType = iota
	secretAuthRequest
	logoutRequest
)

// structs used for json encoding of auth request bodies
type authRequestBody struct {
	Id      string `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	JsonRpc string `json:"jsonrpc"`
}

func buildAuthRequestBody(requestType authRequestType, username, password string) io.Reader {
	var body authRequestBody

	switch requestType {
	case passwordAuthRequest:
		body = authRequestBody{
			Id:     appId,
			Method: "authenticate",
			Params: struct {
				User     string `json:"user"`
				Password string `json:"password"`
				Client   string `json:"client"`
			}{
				User:     username,
				Password: password,
				Client:   appId,
			},
		}
	case secretAuthRequest:
		body = authRequestBody{
			Id:     appId,
			Method: "getUserData2017",
			Params: []struct {
				Auth any `json:"auth"`
			}{{
				Auth: struct {
					ClientTime int64  `json:"clientTime"`
					User       string `json:"user"`
					Otp        string `json:"otp"`
				}{
					ClientTime: time.Now().UnixMilli(),
					User:       username,
					Otp:        password,
				},
			}},
		}
	case logoutRequest:
		body = authRequestBody{
			Id:      appId,
			Method:  "logout",
			Params:  []string{},
			JsonRpc: "2.0",
		}
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
func LoginPassword(username, password string) (session Session, err error) {
	url := baseUrl + "WebUntis/jsonrpc.do?school=" + schoolName
	reqBody := buildAuthRequestBody(passwordAuthRequest, username, password)

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

	return session, nil
}

func LoginSecret(username, secret string, getSessionInfo bool) (session Session, err error) {
	token, _ := totp.GenerateCode(secret, time.Now())
	url := baseUrl + "WebUntis/jsonrpc_intern.do?m=getUserData2017&school=" + schoolName + "&v=i2.2"
	reqBody := buildAuthRequestBody(secretAuthRequest, username, token)

	res, err := http.Post(url, "application/json", reqBody)
	if err != nil {
		return session, err
	}

	sessionId, err := getCookieFromSetCookie(res.Header["Set-Cookie"], "JSESSIONID")
	if err != nil {
		return session, err
	}

	if getSessionInfo {
		// TODO: get all user data
	}

	session = Session{
		SessionId: sessionId,
	}

	return session, nil
}

func getCookieFromSetCookie(header []string, cookieName string) (cookieVal string, err error) {
	if len(header) == 0 {
		return "", errors.New("error: cookie is empty")
	}
	for _, setCookie := range header {
		for _, cookiePart := range strings.Split(setCookie, "; ") {
			cookieSplit := strings.Split(cookiePart, "=")
			if cookieSplit[0] == cookieName {
				return cookieSplit[1], nil
			}
		}
	}
	return "", errors.New("error: cookie not found")
}

func (session *Session) Logout() (err error) {
	url := baseUrl + "WebUntis/jsonrpc.do?school=" + schoolName
	reqBody := buildAuthRequestBody(logoutRequest, "", "")

	_, err = http.Post(url, "application/json", reqBody)
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

	return exams, nil
}

func (session *Session) GetCalendarEvents(start, end time.Time) (events []CalendarEvent, err error) {
	// ensure that external calendars are displayed in timetable
	path := "WebUntis/api/rest/view/v1/timetable/calendar"
	jsonBody := []byte(`{"integrations":[{"name":"Schuljahreskalender","active":true}]}`)
	_, err = session.Request(http.MethodPut, path, nil, jsonBody, true)
	if err != nil {
		return events, err
	}

	// get calendar data by querying timetable
	path = "WebUntis/api/rest/view/v1/timetable/entries"
	queryParams := url.Values{
		"start":        {convertDateToUntis(start)},
		"end":          {convertDateToUntis(end)},
		"format":       {"4"},
		"resourceType": {calendarResourceType},
		"resources":    {strconv.Itoa(calendarResource)},
		"periodTypes":  {"OFFICE_HOUR"}, // often unused period type to query less unneeded information
	}

	res, err := session.Request(http.MethodGet, path, queryParams, nil, true)
	if err != nil {
		return events, err
	}

	var parser fastjson.Parser
	jsonData, err := parser.Parse(res)
	if err != nil {
		return events, err
	}

	events = getCalendarEvents(jsonData)
	return events, nil
}

func getCalendarEvents(jsonData *fastjson.Value) (events []CalendarEvent) {
	// sadly events in "dayEntries" and "gridEntries" have different JSON formats
	for _, dayData := range jsonData.GetArray("days") {
		for _, entry := range dayData.GetArray("dayEntries") {
			if string(entry.GetStringBytes("name")) != "Schuljahreskalender" {
				continue
			}

			entryStart, _ := time.Parse("2006-01-02T15:04", string(entry.GetStringBytes("duration", "start")))
			entryEnd, _ := time.Parse("2006-01-02T15:04:05", string(entry.GetStringBytes("duration", "end")))

			events = append(events, CalendarEvent{
				Id:       entry.GetInt64("id"),
				Name:     string(entry.GetStringBytes("position1", "shortName")),
				Notes:    string(entry.GetStringBytes("notesAll")),
				Date:     string(dayData.GetStringBytes("date")),
				Start:    entryStart,
				End:      entryEnd,
				FullDay:  true,
				Location: string(entry.GetStringBytes("position2", "shortName")),
				Calendar: string(entry.GetStringBytes("position3", "shortName")),
				Color:    string(entry.GetStringBytes("color")),
			})
		}
		for _, entry := range dayData.GetArray("gridEntries") {
			if string(entry.GetStringBytes("name")) != "Schuljahreskalender" {
				continue
			}

			entryStart, _ := time.Parse("2006-01-02T15:04", string(entry.GetStringBytes("duration", "start")))
			entryEnd, _ := time.Parse("2006-01-02T15:04", string(entry.GetStringBytes("duration", "end")))

			events = append(events, CalendarEvent{
				Id:       entry.GetInt64("ids", "0"),
				Name:     string(entry.GetStringBytes("position1", "0", "current", "shortName")),
				Notes:    string(entry.GetStringBytes("notesAll")),
				Date:     string(dayData.GetStringBytes("date")),
				Start:    entryStart,
				End:      entryEnd,
				FullDay:  false,
				Location: string(entry.GetStringBytes("position2", "0", "current", "shortName")),
				Calendar: string(entry.GetStringBytes("position3", "0", "current", "shortName")),
				Color:    string(entry.GetStringBytes("color")),
			})
		}
	}
	return events
}

func (session *Session) GetOverviewEvents(start, end time.Time) (events []TimetableEvent, err error) {
	now, _ := time.Parse("20060102", time.Now().Format("20060102"))
	startDiff := int(start.Sub(now).Hours() / 24)
	endDiff := int(end.Sub(now).Hours()/24) + 1

	for day := startDiff; day <= endDiff; day++ {
		// skip day if date is not available
		if day < 0 || day > 4 {
			continue
		}

		path := "WebUntis/monitor/dayoverview/data?school=" + schoolName
		reqBody := fmt.Sprintf(`{"schoolName": "%s", "format": "Monitor Klassen +%d"}`, schoolName, day)

		res, err := http.Post(baseUrl+path, "application/json", bytes.NewReader([]byte(reqBody)))
		if err != nil {
			return events, err
		}
		defer res.Body.Close()

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return events, err
		}

		var parser fastjson.Parser
		jsonData, err := parser.Parse(string(resBody))
		if err != nil {
			return events, err
		}

		date, _ := time.Parse("20060102", strconv.Itoa(jsonData.GetInt("payload", "date")))
		if date.After(end.Add(24 * time.Hour)) {
			break
		}

		for _, class := range jsonData.GetArray("payload", "rows") {
			className := string(class.GetStringBytes("header"))
			for i, cell := range class.GetArray("cells") {
				if !cell.GetBool("isEvent") {
					continue
				}
				title := string(cell.GetStringBytes("text"))
				start := date.Add(time.Duration(6+i) * time.Hour)
				end := start.Add(time.Duration(cell.GetInt("colSpan")) * time.Hour)

				exists := false
				for i, event := range events {
					if event.Title == title && event.Start.Equal(start) && event.End.Equal(end) {
						events[i].Classes = append(events[i].Classes, className)
						exists = true
					}
				}

				if !exists {
					events = append(events, TimetableEvent{
						Title:   title,
						Classes: []string{className},
						Start:   start,
						End:     end,
					})
				}
			}
		}
	}

	return events, nil
}

func (session *Session) getTeachers() (teachers []UntisValue, err error) {
	path := "WebUntis/api/rest/view/v1/timetable/filter"
	queryParams := url.Values{
		"resourceType":  {"TEACHER"},
		"timetableType": {"STANDARD"},
	}

	res, err := session.Request(http.MethodGet, path, queryParams, nil, true)
	if err != nil {
		return teachers, err
	}

	var parser fastjson.Parser
	jsonData, err := parser.Parse(res)
	if err != nil {
		return teachers, err
	}

	for _, teacher := range jsonData.GetArray("teachers") {
		teachers = append(teachers, UntisValue{
			Id:          teacher.GetInt("teacher", "id"),
			ShortName:   string(teacher.GetStringBytes("teacher", "shortName")),
			LongName:    string(teacher.GetStringBytes("teacher", "longName")),
			DisplayName: string(teacher.GetStringBytes("teacher", "displayName")),
		})
	}

	return teachers, nil
}

func (session *Session) GetTeacherEvents(teacher string, start, end time.Time) (timetableEvents []TimetableEvent, calendarEvents []CalendarEvent, err error) {
	teacherExists := false
	var teacherData UntisValue
	teachers, err := session.getTeachers()
	if err != nil {
		return timetableEvents, calendarEvents, err
	}
	for _, currTeacher := range teachers {
		if currTeacher.DisplayName == teacher {
			teacherExists = true
			teacherData = currTeacher
		}
	}
	if !teacherExists {
		return timetableEvents, calendarEvents, errors.New("error: that teacher does not exist")
	}

	path := "WebUntis/api/rest/view/v1/timetable/entries"
	queryParams := url.Values{
		"start":        {convertDateToUntis(start)},
		"end":          {convertDateToUntis(end)},
		"format":       {"4"},
		"resourceType": {"TEACHER"},
		"resources":    {strconv.Itoa(teacherData.Id)},
		"periodTypes":  {"EVENT"},
	}

	res, err := session.Request(http.MethodGet, path, queryParams, nil, true)
	if err != nil {
		return timetableEvents, calendarEvents, err
	}

	var parser fastjson.Parser
	jsonData, err := parser.Parse(res)
	if err != nil {
		return timetableEvents, calendarEvents, err
	}

	for _, dayData := range jsonData.GetArray("days") {
		for _, entry := range dayData.GetArray("gridEntries") {
			if string(entry.GetStringBytes("type")) != "EVENT" {
				continue
			}

			entryStart, _ := time.Parse("2006-01-02T15:04", string(entry.GetStringBytes("duration", "start")))
			entryEnd, _ := time.Parse("2006-01-02T15:04", string(entry.GetStringBytes("duration", "end")))

			var entryClasses []string
			for _, classData := range entry.GetArray("position1") {
				entryClasses = append(entryClasses, string(classData.GetStringBytes("current", "longName")))
			}

			var entryTeachers []string
			for _, teacherData := range entry.GetArray("position2") {
				entryTeachers = append(entryTeachers, string(teacherData.GetStringBytes("current", "longName")))
			}

			timetableEvents = append(timetableEvents, TimetableEvent{
				Title:    string(entry.GetStringBytes("lessonInfo")),
				Start:    entryStart,
				End:      entryEnd,
				Classes:  entryClasses,
				Teachers: entryTeachers,
			})
		}
	}

	calendarEvents = getCalendarEvents(jsonData)

	return timetableEvents, calendarEvents, nil
}
