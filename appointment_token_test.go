package main

import "testing"

func TestDecideAppointmentAccess(t *testing.T) {
	tests := []struct {
		name    string
		input   tokenRequest
		wantErr bool
		channel string
	}{
		{
			name:    "scopes a token to the appointment room and uses a neutral notice",
			input:   tokenRequest{SessionID: "sess-17", AppointmentID: "apt-42", ClientID: "patient-browser"},
			channel: "appointment-apt-42",
		},
		{
			name:    "rejects a request without a session",
			input:   tokenRequest{AppointmentID: "apt-42", ClientID: "patient-browser"},
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := decideAppointmentAccess(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("decideAppointmentAccess() error = %v, wantErr %v", err, test.wantErr)
			}
			if !test.wantErr && (got.Channel != test.channel || got.Notification != "Appointment chat is ready.") {
				t.Fatalf("decision = %#v", got)
			}
		})
	}
}
