package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ws "github.com/fasthttp/websocket"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pion/webrtc/v4"

	"github.com/FlameInTheDark/gochat/cmd/sfu/config"
	voicev2 "github.com/FlameInTheDark/gochat/cmd/sfu/signaling/v2"
	"github.com/FlameInTheDark/gochat/internal/helper"
	"github.com/FlameInTheDark/gochat/internal/mq/mqmsg"
	"github.com/FlameInTheDark/gochat/internal/permissions"
	"github.com/FlameInTheDark/gochat/internal/shutter"
	"github.com/FlameInTheDark/gochat/internal/voice/dave/wire"
)

type signalTestHarness struct {
	app      *App
	listener net.Listener
	webhook  *httptest.Server
	baseURL  string
}

type gatewayPacket struct {
	Op int             `json:"op"`
	D  json.RawMessage `json:"d"`
}

type v2Client struct {
	conn      *ws.Conn
	sessionID string
	userID    int64
	channelID int64
	token     string
}

var daveKeyPackageFixtures = map[int64]string{
	501: "AAEAAkBBBBoSujDDF1qzYBAnD2jJ2hecTzbOleqEGvvlnPPjJMlb/FMXz9LmD7g4xEeot6IC8N74fhfpN2yGdAPK45esd2tAQQSYZLJhQdGPOfVUtEycQjVRAoq25NBxzaWtqngwkhEthAsFKpIGDX6DAvVS2VUj0O7GtEvuDoMiHcM4YmIRvZfgQEEEQWD9v0QUlBKHzjrNhqX5JduWo7BvOS74KMGD5lCdfH7BYhDQxZum4F2qPO2CTUUAl2b/EOCcdaVZ/lM16RsVpgABCAAAAAAAAAH1AgABAgACAAACAAEBAAAAAAAAAAD//////////wBARzBFAiEA6BtADgVSJ3kr/LwXt4fFVZcBqabPSfqmnA4JqWQz+E0CICnwze3Cserp1dEYe8awDnfTJjqzQeaJlFNJhPJe2qWfAEBHMEUCIQCL2i+ux9W9mrV2KHYvQrf6iFu1sEGi2cf2iqKSf634tQIgNeLXtd1a+szYhtWJXTRTcd+hBxWLAOqn03r9OY4WA4A=",
	502: "AAEAAkBBBDMqRkt2euRV4MrR7y0sWgQNqDfk3NLcK9O/ukaDOkykA1xeWp0r23XQutI9Usy4etMYk/uWWWuRX67nrh6nZPRAQQR3Uln1qePIdu6H4/BhHY1YZAiEdTyKWr0FjM92kjDinkuxemPStmZL5j/qTvd0U+KHlFNWF6rymTCuDt+4n9HrQEEEk7fWAbDxho33M5YHDfzMtxvVInGFW+K6KrD8AJpNOmaA6i9i8TjP6aqNs2v8XQzkVKpAds4Rsbk1rgOR2HXvSAABCAAAAAAAAAH2AgABAgACAAACAAEBAAAAAAAAAAD//////////wBARzBFAiAjcpoWuZiQoVWGNmUVc9thxuNIUzsd5l7QkY9QbRf+KQIhAKAY6Cgv5ACUZ4Mu2ofgsnYVR8KKmkdCgybIXuzTBILYAEBGMEQCIE/2omCFfGN3m2xAaFLkA7bK/UYpH+63jQNyVOU/2remAiAy8V1jYa88NFdE2HoTFCfRDcDhcP7TzKTptfDT/PDsnw==",
	503: "AAEAAkBBBKl+dlrLvNQGPFv6NLwKaTolQpteBzHK6PufVOUhO56P0GttI5QCtjB4jLd0BNL/fYlOlbh34Am/phm7k1CZLvJAQQTb7YceyDEa4uknCPCzKMkqA+XoWQCofDXp/wV7edRRcJhm8+hqe+Ux6gRt8TGIa2FuLLoNgEOzcpVwzhmDEWJQQEEEPcZ0cOJabcQeti08fko3nC1XPmMqIC1LECf6PGnPcGti5Ojtlhw0ONnc7NhtY9LCZoG7xPWVAR2klNHqMbILlAABCAAAAAAAAAH3AgABAgACAAACAAEBAAAAAAAAAAD//////////wBARzBFAiEAzHtWAwhYGPDzdAVY+Kg0YEMNBlX3y1xOKpajWQT8D/cCIGiTR+4MlNkTKGPtvGtpxgX86lmu6eHQWjFnJ6SI/CJVAEBHMEUCIC7sUySxNyMNxrDQKP3gdapJwvnt4QhL4uQxB6yfiNQyAiEA7CDS79o4xF0DGe6b1L2pjVx0XvjOGi0FGnm0sKQj5jI=",
	504: "AAEAAkBBBJFvkKJpG4P7FM/xmqqc+yBD1J39x9oP0d1RRR0JRtx/LSe2CYFV1OLdtYEcHFE+9GgfhD8CIimOd+hzvbfd+apAQQT4duBuDbVZIam4FvsDCtsTmGCk9ctVRrIK3+5pB//IU+r1jk1U6p5xEVShYMnSjdemM0GQhvlHZ0JcAWrA1PqJQEEE8+I9IizpjzYE0OjONzFjIQbye9Grpu7aefSpyAwFq7iZKbiWjXza0vZIlVs9nKASt8xbkJrq7cHom4bJ6LflcQABCAAAAAAAAAH4AgABAgACAAACAAEBAAAAAAAAAAD//////////wBARjBEAiAZuUHGECZlPMK91JSyRiUx9IzxTMru7p6RtTfN20gLkQIgaRvrVf/mkfA3arQClRo9xtBEzCXGlWJufznWWTh1DdYAQEYwRAIgOmWFCZFZ0Ok4kWKYcuOAQIOSZr6vHnxu2itQmoXMTiACIG6oY7yYMN9UmBRzlTvAHOPB3Q7JcUKZ1AldLkuEwdPv",
	801: "AAEAAkBBBFJg5bqn+GCIvXJMOpj9BHmPhEorqPGR/6H+XBpLprFWVpoMQFhpwVKaiyX5Fy+9nORzl0CYf7ZANYMeWVDf+cpAQQQPEuH03FHseRlZPH5qN/OJzyl8Zx4O4z+IuDlPiSc/NgrWHEfn4nWaCo7pwWwgpytQKvLCGyg2JuYvFOf0VEBnQEEEeUJM1L0ytvTl0iWkmdoYLFenn4Hzu4vpGCSzxSAWlC4ppsm/6FargqM2CjLdM1N8fkoMfjR2av/sCtMzz8sj3QABCAAAAAAAAAMhAgABAgACAAACAAEBAAAAAAAAAAD//////////wBARzBFAiEArrXZHzorJ+WYNdXHPEyeDyxcPi4HuNvKQnS2tJzSBzECIC4HXumL1LOp4izaXZbXVSJElj73Qkq9XT7RSypJPB7NAEBHMEUCIQCr4J8QobkRyC1Pksvqzwt6vUDFG8xIfIgCb4FaPlJ56QIgUW1RuWLKopjhL3NF9SE9/3xUi8h4sUApBkCoT76n8VM=",
	802: "AAEAAkBBBCtsCRG1e20d4Ifb+N2dQg1tbxJxXx7i5Pc4StfVz3CJC1Helm8i9y2EGJDX4Sy6jBLEuCyD2JPdV25dd2I9q6ZAQQQaoPB1c4AkgnzZd28UqWlOQ/cOKD2w9F4u1b899QSo/2i8ZfeOQcVgrX+J2JnQkzb6/UiMVKBU98lvlUHxyDKQQEEE0ToVUkXS3LgolrOJv5ARVdLwichW4OcbEOlaKILHkJyKAmeCKmgC1UvoCDn9tKryycsf5Bs4cn3WDWEOpVrgxwABCAAAAAAAAAMiAgABAgACAAACAAEBAAAAAAAAAAD//////////wBASDBGAiEAs3YPp9hBWo7tq7cPNphFST4XBl/aMfgPPp2+ZgYqQBgCIQDe1pOlf5CEjhfNMfXvvI3moj+ZDFzlIS5nOExym/kFSgBARjBEAiBubLVGPoQqGLmzYJPsP+5R9tNrDVJhhv5Aug48fGJ5ogIgHlWmjpgI7i0aNnPTXvJ8ZrTAADDwELgQnGAt+/rEUy0=",
}

func mustFixtureKeyPackage(t *testing.T, userID int64) []byte {
	t.Helper()

	raw, ok := daveKeyPackageFixtures[userID]
	if !ok {
		t.Fatalf("missing DAVE key package fixture for user %d", userID)
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		t.Fatalf("decode key package fixture: %v", err)
	}
	return decoded
}

func newSignalTestHarness(t *testing.T, heartbeatIntervalMS int64) *signalTestHarness {
	t.Helper()

	webhook := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/webhook/sfu/voice/join", "/api/v1/webhook/sfu/voice/leave", "/api/v1/webhook/sfu/channel/alive":
			w.WriteHeader(http.StatusOK)
		case "/api/v1/webhook/sfu/heartbeat":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))

	cfg := &config.Config{
		ServerAddress:             ":0",
		AuthSecret:                "test-secret",
		STUNServers:               []string{"stun:stun.test.invalid:3478"},
		Region:                    "test-region",
		PublicBaseURL:             "http://127.0.0.1",
		WebhookURL:                webhook.URL,
		WebhookToken:              "test-webhook-token",
		ServiceID:                 "test-sfu",
		SignalHeartbeatIntervalMS: heartbeatIntervalMS,
		DAVEEnabled:               true,
		DAVERequiredDefault:       false,
		DAVETransitionTimeoutMS:   100,
		DAVEOldRatchetWindowMS:    10000,
		DAVEAllowAV1:              false,
		MaxAudioBitrateKbps:       64,
		EnforceAudioBitrate:       false,
		AudioBitrateMarginPercent: 15,
	}

	app := NewApp(shutter.NewShutter(newTestLogger()), newTestLogger(), cfg)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.app.Listener(ln)
	}()

	h := &signalTestHarness{
		app:      app,
		listener: ln,
		webhook:  webhook,
		baseURL:  "ws://" + ln.Addr().String() + "/signal",
	}

	t.Cleanup(func() {
		_ = app.Close()
		_ = ln.Close()
		webhook.Close()
		select {
		case err := <-errCh:
			if err != nil && !strings.Contains(err.Error(), "closed") && !strings.Contains(err.Error(), "shutdown") {
				t.Fatalf("listener error: %v", err)
			}
		case <-time.After(time.Second):
		}
	})

	return h
}

func (h *signalTestHarness) dial(t *testing.T, suffix string) *ws.Conn {
	t.Helper()

	dialer := &ws.Dialer{HandshakeTimeout: 5 * time.Second}
	var (
		conn *ws.Conn
		err  error
	)
	for attempt := 0; attempt < 20; attempt++ {
		conn, _, err = dialer.Dial(h.baseURL+suffix, nil)
		if err == nil {
			return conn
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("dial websocket: %v", err)
	return nil
}

func readEnvelopeWithTimeout(t *testing.T, conn *ws.Conn, timeout time.Duration) envelope {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	_, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read message: %v", err)
	}
	var env envelope
	if err := json.Unmarshal(payload, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	return env
}

func waitForEnvelopeType(t *testing.T, conn *ws.Conn, timeout time.Duration, wantType int) envelope {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		env := readEnvelopeWithTimeout(t, conn, time.Until(deadline))
		if env.T == wantType {
			return env
		}
	}
}

func readGatewayPacketWithTimeout(t *testing.T, conn *ws.Conn, timeout time.Duration) gatewayPacket {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	mt, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read gateway packet: %v", err)
	}
	if mt != ws.TextMessage {
		t.Fatalf("expected text gateway packet, got message type %d", mt)
	}
	var packet gatewayPacket
	if err := json.Unmarshal(payload, &packet); err != nil {
		t.Fatalf("unmarshal gateway packet: %v", err)
	}
	return packet
}

func readBinaryMessageWithTimeout(t *testing.T, conn *ws.Conn, timeout time.Duration) []byte {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		_ = conn.SetReadDeadline(deadline)
		mt, payload, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read binary message: %v", err)
		}
		if mt == ws.BinaryMessage {
			return payload
		}
	}
}

func waitForGatewayOp(t *testing.T, conn *ws.Conn, timeout time.Duration, wantOp int) gatewayPacket {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		packet := readGatewayPacketWithTimeout(t, conn, time.Until(deadline))
		if packet.Op == wantOp {
			return packet
		}
	}
}

func waitForBinaryOpcode(t *testing.T, conn *ws.Conn, timeout time.Duration, wantOpcode byte) *wire.DecodedMessage {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		payload := readBinaryMessageWithTimeout(t, conn, time.Until(deadline))
		decoded, err := wire.Decode(payload)
		if err != nil {
			t.Fatalf("decode binary packet: %v", err)
		}
		if decoded.Opcode == wantOpcode {
			return decoded
		}
	}
}

func readNextBinaryDecoded(t *testing.T, conn *ws.Conn, timeout time.Duration) *wire.DecodedMessage {
	t.Helper()

	payload := readBinaryMessageWithTimeout(t, conn, timeout)
	decoded, err := wire.Decode(payload)
	if err != nil {
		t.Fatalf("decode binary packet: %v", err)
	}
	return decoded
}

func writeJSONMessage(t *testing.T, conn *ws.Conn, v any) {
	t.Helper()
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := conn.WriteJSON(v); err != nil {
		t.Fatalf("write json: %v", err)
	}
}

func writeGatewayMessage(t *testing.T, conn *ws.Conn, op int, data any) {
	t.Helper()
	writeJSONMessage(t, conn, voicev2.Packet{Op: op, D: data})
}

func writeBinaryMessage(t *testing.T, conn *ws.Conn, payload []byte) {
	t.Helper()
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := conn.WriteMessage(ws.BinaryMessage, payload); err != nil {
		t.Fatalf("write binary: %v", err)
	}
}

func readCloseCode(t *testing.T, conn *ws.Conn, timeout time.Duration) int {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	_, _, err := conn.ReadMessage()
	if err == nil {
		t.Fatal("expected websocket close")
	}
	var closeErr *ws.CloseError
	if !errors.As(err, &closeErr) {
		t.Fatalf("expected close error, got %v", err)
	}
	return closeErr.Code
}

func issueTestJoinToken(t *testing.T, secret string, userID, channelID int64, perms int64) string {
	t.Helper()
	now := time.Now()
	claims := struct {
		helper.Claims
		ChannelID int64 `json:"channel_id"`
		Perms     int64 `json:"perms"`
	}{
		Claims: helper.Claims{
			UserID:    userID,
			TokenType: "sfu",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "gochat",
				Audience:  []string{"sfu"},
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Minute)),
			},
		},
		ChannelID: channelID,
		Perms:     perms,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func newClientOffer(t *testing.T) (string, *webrtc.PeerConnection) {
	t.Helper()

	pc := newTestPeerConnection(t)
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		t.Fatalf("create offer: %v", err)
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		t.Fatalf("set local description: %v", err)
	}
	select {
	case <-webrtc.GatheringCompletePromise(pc):
	case <-time.After(2 * time.Second):
	}
	if local := pc.LocalDescription(); local != nil {
		return local.SDP, pc
	}
	return offer.SDP, pc
}

func establishV2Participant(t *testing.T, h *signalTestHarness, userID, channelID int64, daveCapable bool) (*v2Client, voicev2.SessionDescription) {
	t.Helper()

	return establishV2ParticipantWithIdentify(t, h, userID, channelID, voicev2.Identify{
		MaxDAVEProtocolVersion:    map[bool]int{true: 1, false: 0}[daveCapable],
		SupportsEncodedTransforms: daveCapable,
		DAVESupported:             daveCapable,
	})
}

func establishV2ParticipantWithIdentify(t *testing.T, h *signalTestHarness, userID, channelID int64, identify voicev2.Identify) (*v2Client, voicev2.SessionDescription) {
	t.Helper()

	perms := int64(permissions.PermVoiceConnect | permissions.PermVoiceSpeak | permissions.PermVoiceVideo)
	token := issueTestJoinToken(t, h.app.cfg.AuthSecret, userID, channelID, perms)
	conn := h.dial(t, "?v=2")

	identify.ChannelID = helper.StringInt64(channelID)
	identify.Token = token
	writeGatewayMessage(t, conn, voicev2.OpIdentify, identify)

	hello := waitForGatewayOp(t, conn, 5*time.Second, voicev2.OpHello)
	var helloPayload voicev2.Hello
	if err := json.Unmarshal(hello.D, &helloPayload); err != nil {
		t.Fatalf("unmarshal hello: %v", err)
	}
	ready := waitForGatewayOp(t, conn, 5*time.Second, voicev2.OpReady)
	var readyPayload voicev2.Ready
	if err := json.Unmarshal(ready.D, &readyPayload); err != nil {
		t.Fatalf("unmarshal ready: %v", err)
	}
	if len(readyPayload.ICEServers) == 0 {
		t.Fatal("expected ice servers in ready payload")
	}

	offerSDP, offerPC := newClientOffer(t)
	defer offerPC.Close()

	writeGatewayMessage(t, conn, voicev2.OpSelectProtocol, voicev2.SelectProtocol{
		Protocol:        "webrtc",
		SDP:             offerSDP,
		Type:            webrtc.SDPTypeOffer.String(),
		RTCConnectionID: fmt.Sprintf("rtc-%d", userID),
	})

	sessionPacket := waitForGatewayOp(t, conn, 5*time.Second, voicev2.OpSessionDescription)
	var sessionDesc voicev2.SessionDescription
	if err := json.Unmarshal(sessionPacket.D, &sessionDesc); err != nil {
		t.Fatalf("unmarshal session description: %v", err)
	}
	if sessionDesc.SDP == "" {
		t.Fatal("expected session description sdp")
	}

	return &v2Client{
		conn:      conn,
		sessionID: helloPayload.SessionID,
		userID:    userID,
		channelID: channelID,
		token:     token,
	}, sessionDesc
}

func TestSignalWSV2MissingMaxDAVEProtocolVersionStillStartsDAVEHandshake(t *testing.T) {
	h := newSignalTestHarness(t, 15000)

	c1, desc1 := establishV2ParticipantWithIdentify(t, h, 801, 140, voicev2.Identify{
		SupportsEncodedTransforms: true,
		DAVESupported:             true,
	})
	defer c1.conn.Close()
	c2, desc2 := establishV2ParticipantWithIdentify(t, h, 802, 140, voicev2.Identify{
		SupportsEncodedTransforms: true,
		DAVESupported:             true,
	})
	defer c2.conn.Close()

	if desc1.DAVEProtocolVersion != 0 || desc2.DAVEProtocolVersion != 0 {
		t.Fatalf("expected transport-only bootstrap before upgrade, got %d and %d", desc1.DAVEProtocolVersion, desc2.DAVEProtocolVersion)
	}

	_ = waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpClientsConnect)
	_ = waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpClientsConnect)

	prepare1 := waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpDAVEPrepareEpoch)
	prepare2 := waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpDAVEPrepareEpoch)
	var payload1 voicev2.PrepareEpoch
	var payload2 voicev2.PrepareEpoch
	if err := json.Unmarshal(prepare1.D, &payload1); err != nil {
		t.Fatalf("unmarshal prepare epoch 1: %v", err)
	}
	if err := json.Unmarshal(prepare2.D, &payload2); err != nil {
		t.Fatalf("unmarshal prepare epoch 2: %v", err)
	}
	if payload1.ProtocolVersion != voicev2.MaxDAVEProtocol || payload2.ProtocolVersion != voicev2.MaxDAVEProtocol {
		t.Fatalf("expected dave prepare epoch protocol %d, got %d and %d", voicev2.MaxDAVEProtocol, payload1.ProtocolVersion, payload2.ProtocolVersion)
	}
}

func completeDAVEUpgrade(t *testing.T, c1, c2 *v2Client) uint16 {
	t.Helper()

	prepareEpoch1 := waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpDAVEPrepareEpoch)
	prepareEpoch2 := waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpDAVEPrepareEpoch)
	var epoch1 voicev2.PrepareEpoch
	var epoch2 voicev2.PrepareEpoch
	_ = json.Unmarshal(prepareEpoch1.D, &epoch1)
	_ = json.Unmarshal(prepareEpoch2.D, &epoch2)
	if epoch1.ProtocolVersion != 1 || epoch2.ProtocolVersion != 1 {
		t.Fatalf("expected protocol version 1 in prepare epoch, got %d and %d", epoch1.ProtocolVersion, epoch2.ProtocolVersion)
	}

	ext1 := waitForBinaryOpcode(t, c1.conn, 5*time.Second, wire.OpcodeExternalSenderPackage)
	ext2 := waitForBinaryOpcode(t, c2.conn, 5*time.Second, wire.OpcodeExternalSenderPackage)
	if ext1.ExternalSender == nil || ext2.ExternalSender == nil {
		t.Fatal("expected external sender packages")
	}

	keyPackage1, _ := wire.EncodeKeyPackage(wire.KeyPackage{Payload: mustFixtureKeyPackage(t, c1.userID)})
	keyPackage2, _ := wire.EncodeKeyPackage(wire.KeyPackage{Payload: mustFixtureKeyPackage(t, c2.userID)})
	writeBinaryMessage(t, c1.conn, keyPackage1)
	writeBinaryMessage(t, c2.conn, keyPackage2)

	props1 := waitForBinaryOpcode(t, c1.conn, 5*time.Second, wire.OpcodeProposals)
	props2 := waitForBinaryOpcode(t, c2.conn, 5*time.Second, wire.OpcodeProposals)
	if len(props1.Payloads) != 1 || len(props2.Payloads) != 1 {
		t.Fatalf("expected one proposals blob per recipient, got %d and %d", len(props1.Payloads), len(props2.Payloads))
	}
	if len(props1.Payloads[0]) == 0 || len(props2.Payloads[0]) == 0 {
		t.Fatal("expected non-empty proposals payloads")
	}

	commit, _ := wire.EncodeCommitWelcome(wire.CommitWelcome{
		Commit:  []byte{11, 12, 13, 14, 15},
		Welcome: []byte{21, 22, 23, 24, 25},
	})
	writeBinaryMessage(t, c1.conn, commit)

	announce1 := waitForBinaryOpcode(t, c1.conn, 5*time.Second, wire.OpcodeAnnounceCommitTransition)
	welcome2 := readNextBinaryDecoded(t, c2.conn, 5*time.Second)
	if announce1.TransitionID == 0 || welcome2.TransitionID == 0 {
		t.Fatal("expected non-zero transition id")
	}
	if welcome2.Opcode != wire.OpcodeWelcome {
		t.Fatalf("expected welcome for non-committer, got opcode %d", welcome2.Opcode)
	}
	if announce1.TransitionID != welcome2.TransitionID {
		t.Fatalf("expected same transition id, got %d and %d", announce1.TransitionID, welcome2.TransitionID)
	}

	writeGatewayMessage(t, c1.conn, voicev2.OpDAVETransitionReady, voicev2.TransitionReady{TransitionID: announce1.TransitionID})
	writeGatewayMessage(t, c2.conn, voicev2.OpDAVETransitionReady, voicev2.TransitionReady{TransitionID: announce1.TransitionID})

	exec1 := waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpDAVEExecuteTransition)
	exec2 := waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpDAVEExecuteTransition)
	var execute1 voicev2.ExecuteTransition
	var execute2 voicev2.ExecuteTransition
	_ = json.Unmarshal(exec1.D, &execute1)
	_ = json.Unmarshal(exec2.D, &execute2)
	if execute1.TransitionID != announce1.TransitionID || execute2.TransitionID != announce1.TransitionID {
		t.Fatalf("execute transition id mismatch: %d %d %d", execute1.TransitionID, execute2.TransitionID, announce1.TransitionID)
	}

	return announce1.TransitionID
}

func TestParseSignalProtocolVersion(t *testing.T) {
	cases := []struct {
		name   string
		raw    string
		want   int
		wantOK bool
	}{
		{name: "default", raw: "", want: signalProtocolVersion1, wantOK: true},
		{name: "v1", raw: "1", want: signalProtocolVersion1, wantOK: true},
		{name: "v2", raw: "2", want: signalProtocolVersion2, wantOK: true},
		{name: "invalid", raw: "3", wantOK: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseSignalProtocolVersion(tc.raw)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if got != tc.want {
				t.Fatalf("version = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestSignalRouteRejectsInvalidVersion(t *testing.T) {
	h := newSignalTestHarness(t, 15000)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com/signal?v=99", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := h.app.app.Test(req)
	if err != nil {
		t.Fatalf("app test: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestSignalWSV1DefaultAndExplicitVersionStillUseLegacyJoinFlow(t *testing.T) {
	h := newSignalTestHarness(t, 15000)
	perms := int64(permissions.PermVoiceConnect | permissions.PermVoiceSpeak | permissions.PermVoiceVideo)

	for _, suffix := range []string{"", "?v=1"} {
		t.Run("suffix="+suffix, func(t *testing.T) {
			conn := h.dial(t, suffix)
			defer conn.Close()

			writeJSONMessage(t, conn, rtcJoinEnvelope{
				OP: int(mqmsg.OPCodeRTC),
				T:  int(mqmsg.EventTypeRTCJoin),
				D: struct {
					Channel helper.StringInt64 `json:"channel"`
					Token   string             `json:"token"`
				}{
					Channel: 42,
					Token:   issueTestJoinToken(t, h.app.cfg.AuthSecret, 101, 42, perms),
				},
			})

			ack := readEnvelopeWithTimeout(t, conn, 5*time.Second)
			if ack.OP != int(mqmsg.OPCodeRTC) || ack.T != int(mqmsg.EventTypeRTCJoin) {
				t.Fatalf("unexpected ack envelope: %+v", ack)
			}

			offer := waitForEnvelopeType(t, conn, 5*time.Second, int(mqmsg.EventTypeRTCOffer))
			if offer.T != int(mqmsg.EventTypeRTCOffer) {
				t.Fatalf("unexpected offer envelope: %+v", offer)
			}
		})
	}
}

func TestSignalWSV1RejectsLegacyJoinWhenDAVEIsRequired(t *testing.T) {
	h := newSignalTestHarness(t, 15000)
	h.app.cfg.DAVERequiredDefault = true

	perms := int64(permissions.PermVoiceConnect | permissions.PermVoiceSpeak | permissions.PermVoiceVideo)
	conn := h.dial(t, "?v=1")
	defer conn.Close()

	writeJSONMessage(t, conn, rtcJoinEnvelope{
		OP: int(mqmsg.OPCodeRTC),
		T:  int(mqmsg.EventTypeRTCJoin),
		D: struct {
			Channel helper.StringInt64 `json:"channel"`
			Token   string             `json:"token"`
		}{
			Channel: 42,
			Token:   issueTestJoinToken(t, h.app.cfg.AuthSecret, 101, 42, perms),
		},
	})

	resp := readEnvelopeWithTimeout(t, conn, 5*time.Second)
	if resp.OP != int(mqmsg.OPCodeRTC) || resp.T != int(mqmsg.EventTypeRTCJoin) {
		t.Fatalf("unexpected response envelope: %+v", resp)
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(resp.D, &errResp); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if errResp.Error != "dave is required; use signal v2" {
		t.Fatalf("error response = %q, want %q", errResp.Error, "dave is required; use signal v2")
	}
}

func TestSignalWSV2NormalCloseFinalizesSessionInsteadOfDetaching(t *testing.T) {
	h := newSignalTestHarness(t, 15000)

	client, _ := establishV2Participant(t, h, 901, 190, true)
	if session := h.app.getSignalV2Session(client.sessionID); session == nil {
		t.Fatalf("expected registered session %q", client.sessionID)
	}

	_ = client.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := client.conn.WriteMessage(ws.CloseMessage, ws.FormatCloseMessage(ws.CloseNormalClosure, "")); err != nil {
		t.Fatalf("write close frame: %v", err)
	}
	_ = client.conn.Close()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if h.app.getSignalV2Session(client.sessionID) == nil {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}

	t.Fatalf("session %q still registered after normal close", client.sessionID)
}

func TestSignalWSV2HandshakeUsesVoiceGatewayOrder(t *testing.T) {
	h := newSignalTestHarness(t, 15000)
	client, sessionDesc := establishV2Participant(t, h, 201, 42, true)
	defer client.conn.Close()

	if sessionDesc.Type != webrtc.SDPTypeAnswer.String() {
		t.Fatalf("session description type = %q, want %q", sessionDesc.Type, webrtc.SDPTypeAnswer.String())
	}
	if sessionDesc.DAVEProtocolVersion != 0 {
		t.Fatalf("initial dave protocol version = %d, want 0", sessionDesc.DAVEProtocolVersion)
	}
	if client.sessionID == "" {
		t.Fatal("expected session id from hello")
	}
}

func TestSignalWSV2HeartbeatTimeoutAfterHello(t *testing.T) {
	h := newSignalTestHarness(t, 20)
	h.app.signalHeartbeatGrace = 20 * time.Millisecond
	perms := int64(permissions.PermVoiceConnect | permissions.PermVoiceSpeak | permissions.PermVoiceVideo)
	conn := h.dial(t, "?v=2")
	defer conn.Close()

	writeGatewayMessage(t, conn, voicev2.OpIdentify, voicev2.Identify{
		ChannelID:                 55,
		Token:                     issueTestJoinToken(t, h.app.cfg.AuthSecret, 301, 55, perms),
		MaxDAVEProtocolVersion:    1,
		SupportsEncodedTransforms: true,
		DAVESupported:             true,
	})
	_ = waitForGatewayOp(t, conn, 5*time.Second, voicev2.OpHello)
	_ = waitForGatewayOp(t, conn, 5*time.Second, voicev2.OpReady)

	if code := readCloseCode(t, conn, 5*time.Second); code != voicev2.CloseCodeHeartbeatTimeout {
		t.Fatalf("close code = %d, want %d", code, voicev2.CloseCodeHeartbeatTimeout)
	}
}

func TestSignalWSV2RejectsInvalidTokenAndLegacyCustomWire(t *testing.T) {
	h := newSignalTestHarness(t, 15000)

	t.Run("invalid token", func(t *testing.T) {
		conn := h.dial(t, "?v=2")
		defer conn.Close()

		writeGatewayMessage(t, conn, voicev2.OpIdentify, voicev2.Identify{
			ChannelID:                 77,
			Token:                     "bad-token",
			MaxDAVEProtocolVersion:    1,
			SupportsEncodedTransforms: true,
		})
		if code := readCloseCode(t, conn, 5*time.Second); code != voicev2.CloseCodeUnauthorized {
			t.Fatalf("close code = %d, want %d", code, voicev2.CloseCodeUnauthorized)
		}
	})

	t.Run("legacy custom wire", func(t *testing.T) {
		conn := h.dial(t, "?v=2")
		defer conn.Close()

		writeJSONMessage(t, conn, map[string]any{
			"op": 7,
			"t":  530,
			"d": map[string]any{
				"channel": 55,
				"token":   "legacy",
			},
		})
		if code := readCloseCode(t, conn, 5*time.Second); code != voicev2.CloseCodeInvalidPayload {
			t.Fatalf("close code = %d, want %d", code, voicev2.CloseCodeInvalidPayload)
		}
	})
}

func TestSignalWSV2RejectsWrongPhaseUnsupportedProtocolAndMalformedSDP(t *testing.T) {
	h := newSignalTestHarness(t, 15000)

	t.Run("wrong phase", func(t *testing.T) {
		conn := h.dial(t, "?v=2")
		defer conn.Close()

		writeGatewayMessage(t, conn, voicev2.OpSelectProtocol, voicev2.SelectProtocol{
			Protocol:        "webrtc",
			SDP:             "v=0",
			RTCConnectionID: "rtc-wrong-phase",
		})
		if code := readCloseCode(t, conn, 5*time.Second); code != voicev2.CloseCodeWrongPhase {
			t.Fatalf("close code = %d, want %d", code, voicev2.CloseCodeWrongPhase)
		}
	})

	t.Run("unsupported protocol", func(t *testing.T) {
		client, _ := establishV2Participant(t, h, 401, 88, true)
		defer client.conn.Close()

		writeGatewayMessage(t, client.conn, voicev2.OpSelectProtocol, voicev2.SelectProtocol{
			Protocol:        "udp",
			SDP:             "v=0",
			RTCConnectionID: "rtc-bad-protocol",
			Type:            webrtc.SDPTypeAnswer.String(),
		})
		if code := readCloseCode(t, client.conn, 5*time.Second); code != voicev2.CloseCodeUnsupportedMedium {
			t.Fatalf("close code = %d, want %d", code, voicev2.CloseCodeUnsupportedMedium)
		}
	})

	t.Run("malformed sdp", func(t *testing.T) {
		perms := int64(permissions.PermVoiceConnect | permissions.PermVoiceSpeak | permissions.PermVoiceVideo)
		conn := h.dial(t, "?v=2")
		defer conn.Close()

		writeGatewayMessage(t, conn, voicev2.OpIdentify, voicev2.Identify{
			ChannelID:                 89,
			Token:                     issueTestJoinToken(t, h.app.cfg.AuthSecret, 402, 89, perms),
			MaxDAVEProtocolVersion:    1,
			SupportsEncodedTransforms: true,
			DAVESupported:             true,
		})
		_ = waitForGatewayOp(t, conn, 5*time.Second, voicev2.OpHello)
		_ = waitForGatewayOp(t, conn, 5*time.Second, voicev2.OpReady)

		writeGatewayMessage(t, conn, voicev2.OpSelectProtocol, voicev2.SelectProtocol{
			Protocol:        "webrtc",
			SDP:             "not-an-sdp",
			Type:            webrtc.SDPTypeOffer.String(),
			RTCConnectionID: "rtc-bad-sdp",
		})
		if code := readCloseCode(t, conn, 5*time.Second); code != voicev2.CloseCodeInvalidPayload {
			t.Fatalf("close code = %d, want %d", code, voicev2.CloseCodeInvalidPayload)
		}
	})
}

func TestSignalWSV2InvalidCommitWelcomeTriggersRecreate(t *testing.T) {
	h := newSignalTestHarness(t, 15000)

	c1, desc1 := establishV2Participant(t, h, 501, 100, true)
	defer c1.conn.Close()
	c2, desc2 := establishV2Participant(t, h, 502, 100, true)
	defer c2.conn.Close()

	if desc1.DAVEProtocolVersion != 0 || desc2.DAVEProtocolVersion != 0 {
		t.Fatalf("expected transport-only bootstrap, got %d and %d", desc1.DAVEProtocolVersion, desc2.DAVEProtocolVersion)
	}

	clientsConnect1 := waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpClientsConnect)
	var connectPayload1 voicev2.ClientsConnect
	_ = json.Unmarshal(clientsConnect1.D, &connectPayload1)
	if len(connectPayload1.UserIDs) != 1 || connectPayload1.UserIDs[0] != "502" {
		t.Fatalf("unexpected clients_connect payload for c1: %+v", connectPayload1)
	}
	clientsConnect2 := waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpClientsConnect)
	var connectPayload2 voicev2.ClientsConnect
	_ = json.Unmarshal(clientsConnect2.D, &connectPayload2)
	if len(connectPayload2.UserIDs) != 1 || connectPayload2.UserIDs[0] != "501" {
		t.Fatalf("unexpected clients_connect payload for c2: %+v", connectPayload2)
	}

	lastTransitionID := completeDAVEUpgrade(t, c1, c2)

	writeGatewayMessage(t, c1.conn, voicev2.OpDAVEInvalidCommitWelcome, voicev2.InvalidCommitWelcome{TransitionID: lastTransitionID})
	_ = waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpDAVEPrepareEpoch)
	_ = waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpDAVEPrepareEpoch)
	_ = waitForBinaryOpcode(t, c1.conn, 5*time.Second, wire.OpcodeExternalSenderPackage)
	_ = waitForBinaryOpcode(t, c2.conn, 5*time.Second, wire.OpcodeExternalSenderPackage)
}

func TestSignalWSV2DAVEUpgradeLateJoinResumeAndDowngrade(t *testing.T) {
	h := newSignalTestHarness(t, 15000)

	c1, desc1 := establishV2Participant(t, h, 501, 100, true)
	defer c1.conn.Close()
	c2, desc2 := establishV2Participant(t, h, 502, 100, true)
	defer c2.conn.Close()

	if desc1.DAVEProtocolVersion != 0 || desc2.DAVEProtocolVersion != 0 {
		t.Fatalf("expected transport-only bootstrap, got %d and %d", desc1.DAVEProtocolVersion, desc2.DAVEProtocolVersion)
	}

	_ = waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpClientsConnect)
	_ = waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpClientsConnect)
	_ = completeDAVEUpgrade(t, c1, c2)

	c3, desc3 := establishV2Participant(t, h, 503, 100, true)
	defer c3.conn.Close()
	if desc3.DAVEProtocolVersion != 0 {
		t.Fatalf("late dave join should bootstrap in transport mode, got %d", desc3.DAVEProtocolVersion)
	}

	_ = waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpClientsConnect)
	_ = waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpClientsConnect)
	_ = waitForGatewayOp(t, c3.conn, 5*time.Second, voicev2.OpClientsConnect)
	prepare1 := waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpDAVEPrepareEpoch)
	prepare2 := waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpDAVEPrepareEpoch)
	prepare3 := waitForGatewayOp(t, c3.conn, 5*time.Second, voicev2.OpDAVEPrepareEpoch)
	var epoch1 voicev2.PrepareEpoch
	var epoch2 voicev2.PrepareEpoch
	var epoch3 voicev2.PrepareEpoch
	_ = json.Unmarshal(prepare1.D, &epoch1)
	_ = json.Unmarshal(prepare2.D, &epoch2)
	_ = json.Unmarshal(prepare3.D, &epoch3)
	if epoch1.ProtocolVersion != 1 || epoch2.ProtocolVersion != 1 || epoch3.ProtocolVersion != 1 {
		t.Fatalf("expected recreate protocol version 1, got %+v %+v %+v", epoch1, epoch2, epoch3)
	}
	if epoch1.Epoch != 2 || epoch2.Epoch != 2 || epoch3.Epoch != 2 {
		t.Fatalf("expected recreate epoch 2, got %+v %+v %+v", epoch1, epoch2, epoch3)
	}
	_ = waitForBinaryOpcode(t, c1.conn, 5*time.Second, wire.OpcodeExternalSenderPackage)
	_ = waitForBinaryOpcode(t, c2.conn, 5*time.Second, wire.OpcodeExternalSenderPackage)
	_ = waitForBinaryOpcode(t, c3.conn, 5*time.Second, wire.OpcodeExternalSenderPackage)

	keyPackage1, _ := wire.EncodeKeyPackage(wire.KeyPackage{Payload: mustFixtureKeyPackage(t, c1.userID)})
	keyPackage2, _ := wire.EncodeKeyPackage(wire.KeyPackage{Payload: mustFixtureKeyPackage(t, c2.userID)})
	keyPackage3, _ := wire.EncodeKeyPackage(wire.KeyPackage{Payload: mustFixtureKeyPackage(t, c3.userID)})
	writeBinaryMessage(t, c1.conn, keyPackage1)
	writeBinaryMessage(t, c2.conn, keyPackage2)
	writeBinaryMessage(t, c3.conn, keyPackage3)

	props1 := waitForBinaryOpcode(t, c1.conn, 5*time.Second, wire.OpcodeProposals)
	props2 := waitForBinaryOpcode(t, c2.conn, 5*time.Second, wire.OpcodeProposals)
	props3 := waitForBinaryOpcode(t, c3.conn, 5*time.Second, wire.OpcodeProposals)
	if len(props1.Payloads) != 1 || len(props2.Payloads) != 1 || len(props3.Payloads) != 1 {
		t.Fatalf("expected one proposals blob per recipient, got %d, %d, %d", len(props1.Payloads), len(props2.Payloads), len(props3.Payloads))
	}
	if len(props1.Payloads[0]) == 0 || len(props2.Payloads[0]) == 0 || len(props3.Payloads[0]) == 0 {
		t.Fatal("expected non-empty proposals payloads for late join upgrade")
	}

	commit3, _ := wire.EncodeCommitWelcome(wire.CommitWelcome{
		Commit:  []byte{61, 62, 63, 64, 65},
		Welcome: []byte{71, 72, 73, 74, 75},
	})
	writeBinaryMessage(t, c1.conn, commit3)

	announce3 := waitForBinaryOpcode(t, c1.conn, 5*time.Second, wire.OpcodeAnnounceCommitTransition)
	welcomeC2 := readNextBinaryDecoded(t, c2.conn, 5*time.Second)
	welcomeC3 := readNextBinaryDecoded(t, c3.conn, 5*time.Second)
	if welcomeC2.Opcode != wire.OpcodeWelcome || welcomeC3.Opcode != wire.OpcodeWelcome {
		t.Fatalf("expected welcome for recreate recipients, got opcodes %d and %d", welcomeC2.Opcode, welcomeC3.Opcode)
	}
	if welcomeC2.TransitionID != announce3.TransitionID || welcomeC3.TransitionID != announce3.TransitionID {
		t.Fatalf("expected matching transition ids, got %d, %d, %d", announce3.TransitionID, welcomeC2.TransitionID, welcomeC3.TransitionID)
	}
	writeGatewayMessage(t, c1.conn, voicev2.OpDAVETransitionReady, voicev2.TransitionReady{TransitionID: announce3.TransitionID})
	writeGatewayMessage(t, c2.conn, voicev2.OpDAVETransitionReady, voicev2.TransitionReady{TransitionID: announce3.TransitionID})
	writeGatewayMessage(t, c3.conn, voicev2.OpDAVETransitionReady, voicev2.TransitionReady{TransitionID: announce3.TransitionID})
	_ = waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpDAVEExecuteTransition)
	_ = waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpDAVEExecuteTransition)
	_ = waitForGatewayOp(t, c3.conn, 5*time.Second, voicev2.OpDAVEExecuteTransition)

	oldConn := c1.conn
	if err := oldConn.Close(); err != nil {
		t.Fatalf("close old conn: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	resumedConn := h.dial(t, "?v=2")
	writeGatewayMessage(t, resumedConn, voicev2.OpResume, voicev2.Resume{
		SessionID: c1.sessionID,
		ChannelID: helper.StringInt64(c1.channelID),
		Token:     c1.token,
	})
	resumedHello := waitForGatewayOp(t, resumedConn, 5*time.Second, voicev2.OpHello)
	var helloPayload voicev2.Hello
	_ = json.Unmarshal(resumedHello.D, &helloPayload)
	if helloPayload.SessionID != c1.sessionID {
		t.Fatalf("resumed hello session id = %q, want %q", helloPayload.SessionID, c1.sessionID)
	}
	resumed := waitForGatewayOp(t, resumedConn, 5*time.Second, voicev2.OpResumed)
	var resumedPayload voicev2.Resumed
	_ = json.Unmarshal(resumed.D, &resumedPayload)
	if resumedPayload.DAVEProtocolVersion != 1 {
		t.Fatalf("resumed dave protocol version = %d, want 1", resumedPayload.DAVEProtocolVersion)
	}
	c1.conn = resumedConn
	defer resumedConn.Close()

	c4, desc4 := establishV2Participant(t, h, 504, 100, false)
	defer c4.conn.Close()
	if desc4.DAVEProtocolVersion != 0 {
		t.Fatalf("non-dave join should receive transport-only bootstrap, got %d", desc4.DAVEProtocolVersion)
	}

	_ = waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpClientsConnect)
	_ = waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpClientsConnect)
	_ = waitForGatewayOp(t, c3.conn, 5*time.Second, voicev2.OpClientsConnect)
	_ = waitForGatewayOp(t, c4.conn, 5*time.Second, voicev2.OpClientsConnect)

	prepareTransition1 := waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpDAVEPrepareTransition)
	prepareTransition2 := waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpDAVEPrepareTransition)
	prepareTransition3 := waitForGatewayOp(t, c3.conn, 5*time.Second, voicev2.OpDAVEPrepareTransition)
	var down1 voicev2.PrepareTransition
	var down2 voicev2.PrepareTransition
	var down3 voicev2.PrepareTransition
	_ = json.Unmarshal(prepareTransition1.D, &down1)
	_ = json.Unmarshal(prepareTransition2.D, &down2)
	_ = json.Unmarshal(prepareTransition3.D, &down3)
	if down1.ProtocolVersion != 0 || down2.ProtocolVersion != 0 || down3.ProtocolVersion != 0 {
		t.Fatalf("expected downgrade to protocol version 0, got %+v %+v %+v", down1, down2, down3)
	}
	writeGatewayMessage(t, c1.conn, voicev2.OpDAVETransitionReady, voicev2.TransitionReady{TransitionID: down1.TransitionID})
	writeGatewayMessage(t, c2.conn, voicev2.OpDAVETransitionReady, voicev2.TransitionReady{TransitionID: down1.TransitionID})
	writeGatewayMessage(t, c3.conn, voicev2.OpDAVETransitionReady, voicev2.TransitionReady{TransitionID: down1.TransitionID})
	_ = waitForGatewayOp(t, c1.conn, 5*time.Second, voicev2.OpDAVEExecuteTransition)
	_ = waitForGatewayOp(t, c2.conn, 5*time.Second, voicev2.OpDAVEExecuteTransition)
	_ = waitForGatewayOp(t, c3.conn, 5*time.Second, voicev2.OpDAVEExecuteTransition)
}
