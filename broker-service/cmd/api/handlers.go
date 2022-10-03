package main

import (
	"broker/event"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {

	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}
	app.writeJson(w, http.StatusOK, payload)

}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {

	var requestPayload RequestPayload

	err := app.readJson(w, r, &requestPayload)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.Authenticate(w, requestPayload.Auth)
	case "log":
		app.logEventviaRabbit(w, requestPayload.Log)
	case "mail":
		app.SendMail(w, requestPayload.Mail)
	default:

	}
	// app.writeJson(w, http.StatusOK, payload)

}

func (app *Config) Authenticate(w http.ResponseWriter, a AuthPayload) {

	jsonData, _ := json.MarshalIndent(a, "", "\t")

	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJson(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJson(w, errors.New("invalid creds"))
		return
	} else if response.StatusCode != http.StatusOK {

		app.errorJson(w, errors.New("error calling auth service"), response.StatusCode)
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	if jsonFromService.Error {
		app.errorJson(w, err, http.StatusUnauthorized)
		return
	}
	var payload jsonResponse
	payload.Error = false
	payload.Message = "authenticated"
	payload.Data = jsonFromService.Data

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) LogItem(w http.ResponseWriter, log LogPayload) {
	jsonData, _ := json.MarshalIndent(log, "", "\t")

	logServiceUrl := "http://logger-service/log"
	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJson(w, err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {

		app.errorJson(w, errors.New("error calling log service"), response.StatusCode)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) SendMail(w http.ResponseWriter, mail MailPayload) {
	jsonData, _ := json.MarshalIndent(mail, "", "\t")

	mailServiceUrl := "http://mail-service/mail"
	request, err := http.NewRequest("POST", mailServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJson(w, err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {

		app.errorJson(w, errors.New("error calling mail service"), response.StatusCode)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) logEventviaRabbit(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged via msg"

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name string, msg string) error {
	emitter, err := event.NewEventEmitter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: msg,
	}

	j, _ := json.MarshalIndent(&payload, "", "")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}

	return nil
}

func (app *Config) QueueLogItem(w http.ResponseWriter, log LogPayload) {
	jsonData, _ := json.MarshalIndent(log, "", "\t")

	logServiceUrl := "http://logger-service/log"
	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJson(w, err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {

		app.errorJson(w, errors.New("error calling log service"), response.StatusCode)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "logged"

	app.writeJson(w, http.StatusAccepted, payload)
}
