package wire

import (
	"encoding/hex"
	"testing"
)

func mustHex(t *testing.T, raw []byte) string {
	t.Helper()
	return hex.EncodeToString(raw)
}

func TestBinaryCodecs_GoldenBytes(t *testing.T) {
	tests := []struct {
		name string
		got  []byte
		want string
	}{
		{
			name: "external sender package",
			got: mustEncodeFunc(t, func() ([]byte, error) {
				return EncodeExternalSenderPackage(ExternalSenderPackage{
					SequenceNumber: 1,
					ExternalSender: ExternalSender{
						SignatureKey:   []byte{0x01, 0x02},
						CredentialType: CredentialTypeBasic,
						Identity:       []byte{0x03, 0x04},
					},
				})
			}),
			want: "0001190201020001020304",
		},
		{
			name: "key package",
			got:  mustEncodeFunc(t, func() ([]byte, error) { return EncodeKeyPackage(KeyPackage{Payload: []byte{0xaa, 0xbb, 0xcc}}) }),
			want: "1a03aabbcc",
		},
		{
			name: "proposals append",
			got: mustEncodeFunc(t, func() ([]byte, error) {
				return EncodeProposals(Proposals{
					SequenceNumber:   2,
					OperationType:    ProposalsAppend,
					ProposalMessages: [][]byte{{0x01}, {0x02, 0x03}},
				})
			}),
			want: "00021b00050101020203",
		},
		{
			name: "commit welcome",
			got: mustEncodeFunc(t, func() ([]byte, error) {
				return EncodeCommitWelcome(CommitWelcome{
					Commit:  []byte{0x10, 0x11},
					Welcome: []byte{0x20, 0x21, 0x22},
				})
			}),
			want: "1c02101103202122",
		},
		{
			name: "announce commit transition",
			got: mustEncodeFunc(t, func() ([]byte, error) {
				return EncodeAnnounceCommitTransition(AnnounceCommitTransition{
					SequenceNumber: 3,
					TransitionID:   9,
					Commit:         []byte{0x30, 0x31},
				})
			}),
			want: "00031d0009023031",
		},
		{
			name: "welcome",
			got: mustEncodeFunc(t, func() ([]byte, error) {
				return EncodeWelcome(Welcome{
					SequenceNumber: 4,
					TransitionID:   9,
					Welcome:        []byte{0x40, 0x41},
				})
			}),
			want: "00041e0009024041",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := mustHex(t, tc.got); got != tc.want {
				t.Fatalf("hex = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestBinaryCodecs_DecodeRoundTrip(t *testing.T) {
	payloads := [][]byte{
		mustEncodeFunc(t, func() ([]byte, error) {
			return EncodeExternalSenderPackage(ExternalSenderPackage{
				SequenceNumber: 7,
				ExternalSender: ExternalSender{
					SignatureKey:   []byte{0x01, 0x02, 0x03},
					CredentialType: CredentialTypeBasic,
					Identity:       []byte{0x04, 0x05},
				},
			})
		}),
		mustEncodeFunc(t, func() ([]byte, error) { return EncodeKeyPackage(KeyPackage{Payload: []byte{0x10, 0x11, 0x12, 0x13}}) }),
		mustEncodeFunc(t, func() ([]byte, error) {
			return EncodeProposals(Proposals{
				SequenceNumber:   8,
				OperationType:    ProposalsAppend,
				ProposalMessages: [][]byte{{0x20, 0x21}, {0x22}},
			})
		}),
		mustEncodeFunc(t, func() ([]byte, error) {
			return EncodeCommitWelcome(CommitWelcome{
				Commit:  []byte{0x30, 0x31},
				Welcome: []byte{0x32, 0x33},
			})
		}),
		mustEncodeFunc(t, func() ([]byte, error) {
			return EncodeAnnounceCommitTransition(AnnounceCommitTransition{
				SequenceNumber: 9,
				TransitionID:   2,
				Commit:         []byte{0x40, 0x41},
			})
		}),
		mustEncodeFunc(t, func() ([]byte, error) {
			return EncodeWelcome(Welcome{
				SequenceNumber: 10,
				TransitionID:   2,
				Welcome:        []byte{0x50, 0x51},
			})
		}),
	}

	for _, payload := range payloads {
		decoded, err := Decode(payload)
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if decoded.Opcode == 0 {
			t.Fatal("expected opcode")
		}
	}
}

func mustEncodeFunc(t *testing.T, fn func() ([]byte, error)) []byte {
	t.Helper()
	raw, err := fn()
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return raw
}
