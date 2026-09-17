package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
)

type chatGateway struct {
	infrai *infraiClient
}

type tokenResponse struct {
	Channel      string `json:"channel"`
	Token        string `json:"token"`
	Notification string `json:"notification"`
}

func main() {
	client, err := newInfraiClient()
	if err != nil {
		log.Fatal(err)
	}
	gateway := chatGateway{infrai: client}
	http.HandleFunc("/appointment-token", gateway.handleAppointmentToken)
	address := os.Getenv("ADDR")
	if address == "" {
		address = ":8080"
	}
	log.Printf("appointment chat gateway listening on %s", address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func (g chatGateway) handleAppointmentToken(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer request.Body.Close()
	var input tokenRequest
	if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	decision, err := decideAppointmentAccess(input)
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := g.infrai.verifySession(request.Context(), input.SessionID); err != nil {
		writeInfraiError(writer, err)
		return
	}
	issued, err := g.infrai.issueChannelToken(request.Context(), issueTokenInput{
		ClientID: input.ClientID, Channels: []string{decision.Channel}, Capabilities: decision.Capabilities, TTLSeconds: 900,
	})
	if err != nil {
		writeInfraiError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, tokenResponse{Channel: decision.Channel, Token: issued.Token, Notification: decision.Notification})
}

func writeInfraiError(writer http.ResponseWriter, err error) {
	status := http.StatusBadGateway
	var rejected *infraiError
	if errors.As(err, &rejected) {
		status = http.StatusBadRequest
	}
	writeJSON(writer, status, map[string]string{"error": strings.TrimSpace(err.Error())})
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
