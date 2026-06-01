package goclient

import "testing"

func TestDefaultEndpointsUseHostedGoChat(t *testing.T) {
	session, err := New("gcb_test")
	if err != nil {
		t.Fatal(err)
	}
	if session.APIEndpoint != "https://gochat.anticode.dev" {
		t.Fatalf("api endpoint = %q", session.APIEndpoint)
	}
	if session.GatewayEndpoint != "wss://gochat.anticode.dev/bot/ws" {
		t.Fatalf("gateway endpoint = %q", session.GatewayEndpoint)
	}
}

func TestWithEndpointConfiguresAPIAndGateway(t *testing.T) {
	session, err := New("gcb_test", WithEndpoint("http://127.0.0.1:8080/base/"))
	if err != nil {
		t.Fatal(err)
	}
	if session.APIEndpoint != "http://127.0.0.1:8080/base" {
		t.Fatalf("api endpoint = %q", session.APIEndpoint)
	}
	if session.GatewayEndpoint != "ws://127.0.0.1:8080/base/bot/ws" {
		t.Fatalf("gateway endpoint = %q", session.GatewayEndpoint)
	}
}
