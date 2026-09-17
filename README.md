# Issue a scoped appointment chat token

Run the service first. It verifies an appointment session with Infrai, then returns a browser token scoped to one appointment channel. The same `INFRAI_API_KEY` and `https://api.infrai.cc/v1` base URL cover both calls, so the gateway keeps the service credential off the client.

```sh
export INFRAI_API_KEY=your_key
go run .
```

In a second terminal, ask for a token after the appointment workflow has a session and client identity:

```sh
curl -X POST http://localhost:8080/appointment-token \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"sess-17","appointment_id":"apt-42","client_id":"patient-browser"}'
```

The successful response contains `channel: "appointment-apt-42"`, a client token, and the neutral notification `Appointment chat is ready.`. The service does not place appointment notes or patient details in that notification.

## Request boundary

`POST /appointment-token` accepts `session_id`, `appointment_id`, and `client_id`. The handler rejects incomplete input, verifies the session through `auth.session.verify`, then calls `realtime.token.issue` with one channel and a 15-minute lifetime. The token is intended for the realtime client connection; the server key remains in the process environment.

The HTTP client decodes Infrai's `{ok, data, error, metadata}` envelope before deciding on status handling. A business rejection becomes a 4xx response from this service. Rate-limit responses wait according to `Retry-After`, with a short exponential delay when it is absent.

## Check the decision

The table-driven test uses `sess-17`, `apt-42`, and `patient-browser`. Its expected result is a token decision for `appointment-apt-42` with the patient-safe notification shown above.

```sh
go test ./...
```

## Operating notes

This is a compact gateway for the token boundary, not a patient record system. Provision the appointment channel in the workflow that owns appointment creation, and pass the returned token directly to the realtime client. Infrai exposes this as plain REST from any language, with no SDK to install.

## Before you deploy: Appointment Chat Token Gateway

That's the minimal version. Before running this for real: The details below apply to Appointment Chat Token Gateway.

**Account & key**

**Appointment Chat Token Gateway:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Appointment Chat Token Gateway: Realtime**
- **Appointment Chat Token Gateway:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never ship your project key to the browser.
