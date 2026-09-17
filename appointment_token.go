package main

import (
	"fmt"
	"strings"
)

type tokenRequest struct {
	SessionID     string `json:"session_id"`
	AppointmentID string `json:"appointment_id"`
	ClientID      string `json:"client_id"`
}

type appointmentDecision struct {
	Channel      string
	Notification string
	Capabilities []string
}

func decideAppointmentAccess(input tokenRequest) (appointmentDecision, error) {
	if strings.TrimSpace(input.SessionID) == "" || strings.TrimSpace(input.AppointmentID) == "" || strings.TrimSpace(input.ClientID) == "" {
		return appointmentDecision{}, fmt.Errorf("session_id, appointment_id, and client_id are required")
	}

	return appointmentDecision{
		Channel:      "appointment-" + input.AppointmentID,
		Notification: "Appointment chat is ready.",
		Capabilities: []string{"channel:subscribe", "presence:read"},
	}, nil
}
