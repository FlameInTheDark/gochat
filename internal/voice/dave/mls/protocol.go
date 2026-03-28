package mls

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

const (
	protocolVersionMLS10      = 1
	wireFormatPublicMessage   = 0x0001
	wireFormatWelcome         = 0x0003
	wireFormatKeyPackage      = 0x0005
	contentTypeProposal       = 0x02
	senderTypeExternal        = 0x02
	proposalTypeAdd           = 0x0001
	credentialTypeBasic       = 0x0001
	daveCipherSuiteP256SHA256 = 0x0002
)

type KeyPackage struct {
	Raw          []byte
	Inner        []byte
	UserID       int64
	Identity     []byte
	SignatureKey []byte
}

type ExternalSender struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
	Identity   []byte
}

func ValidateKeyPackage(raw []byte) error {
	if len(raw) == 0 {
		return fmt.Errorf("empty key package")
	}
	if len(raw) < 4 {
		return fmt.Errorf("key package too short")
	}
	if _, err := parseKeyPackage(raw); err != nil {
		return err
	}
	return nil
}

func ValidateCommit(raw []byte) error {
	if len(raw) == 0 {
		return fmt.Errorf("empty commit")
	}
	if len(raw) < 4 {
		return fmt.Errorf("commit too short")
	}
	return nil
}

func ValidateWelcome(raw []byte) error {
	if len(raw) == 0 {
		return fmt.Errorf("empty welcome")
	}
	if len(raw) < 4 {
		return fmt.Errorf("welcome too short")
	}
	return nil
}

func ParseAndValidateKeyPackage(raw []byte, expectedUserID int64) (*KeyPackage, error) {
	kp, err := parseKeyPackage(raw)
	if err != nil {
		return nil, err
	}
	if expectedUserID != 0 && kp.UserID != expectedUserID {
		return nil, fmt.Errorf("key package user id %d does not match authenticated user %d", kp.UserID, expectedUserID)
	}
	return kp, nil
}

func NewExternalSender(identity []byte) (*ExternalSender, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return &ExternalSender{
		PrivateKey: priv,
		PublicKey:  elliptic.Marshal(elliptic.P256(), priv.PublicKey.X, priv.PublicKey.Y),
		Identity:   append([]byte(nil), identity...),
	}, nil
}

func GroupIDForChannel(channelID int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(channelID))
	return buf
}

func BuildExternalAddProposal(channelID int64, epoch uint64, senderIndex uint32, sender *ExternalSender, keyPackage *KeyPackage) ([]byte, []byte, error) {
	if sender == nil || sender.PrivateKey == nil {
		return nil, nil, fmt.Errorf("external sender is not initialized")
	}
	if keyPackage == nil || len(keyPackage.Inner) == 0 {
		return nil, nil, fmt.Errorf("key package is empty")
	}

	framed := make([]byte, 0, 256+len(keyPackage.Inner))
	framed = appendOpaque(framed, GroupIDForChannel(channelID))
	framed = binary.BigEndian.AppendUint64(framed, epoch)
	framed = append(framed, senderTypeExternal)
	framed = binary.BigEndian.AppendUint32(framed, senderIndex)
	framed = appendOpaque(framed, nil) // authenticated_data
	framed = append(framed, contentTypeProposal)
	framed = binary.BigEndian.AppendUint16(framed, proposalTypeAdd)
	framed = append(framed, keyPackage.Inner...)

	tbs := make([]byte, 0, 4+len(framed))
	tbs = binary.BigEndian.AppendUint16(tbs, protocolVersionMLS10)
	tbs = binary.BigEndian.AppendUint16(tbs, wireFormatPublicMessage)
	tbs = append(tbs, framed...)

	signature, err := signWithLabel(sender.PrivateKey, "FramedContentTBS", tbs)
	if err != nil {
		return nil, nil, err
	}
	authData := appendOpaque(nil, signature)

	authenticatedContent := make([]byte, 0, 2+len(framed)+len(authData))
	authenticatedContent = binary.BigEndian.AppendUint16(authenticatedContent, wireFormatPublicMessage)
	authenticatedContent = append(authenticatedContent, framed...)
	authenticatedContent = append(authenticatedContent, authData...)

	ref, err := refHash("MLS 1.0 Proposal Reference", authenticatedContent)
	if err != nil {
		return nil, nil, err
	}

	message := make([]byte, 0, 4+len(framed)+len(authData))
	message = binary.BigEndian.AppendUint16(message, protocolVersionMLS10)
	message = binary.BigEndian.AppendUint16(message, wireFormatPublicMessage)
	message = append(message, framed...)
	message = append(message, authData...)
	return message, ref, nil
}

func parseKeyPackage(raw []byte) (_ *KeyPackage, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("invalid key package: %v", v)
			}
		}
	}()

	if len(raw) < 4 {
		return nil, fmt.Errorf("key package too short")
	}
	inner := raw
	version := binary.BigEndian.Uint16(inner[:2])
	if version != protocolVersionMLS10 {
		return nil, fmt.Errorf("unsupported key package protocol version %d", version)
	}
	if binary.BigEndian.Uint16(inner[2:4]) == wireFormatKeyPackage {
		inner = raw[4:]
		if len(inner) < 4 {
			return nil, fmt.Errorf("key package payload too short")
		}
	}
	offset := 0

	if readUint16(inner, &offset) != protocolVersionMLS10 {
		return nil, fmt.Errorf("unsupported inner key package version")
	}
	if cs := readUint16(inner, &offset); cs != daveCipherSuiteP256SHA256 {
		return nil, fmt.Errorf("unsupported DAVE key package ciphersuite %d", cs)
	}
	if _, err := readOpaque(inner, &offset); err != nil {
		return nil, err
	}

	leafNodeStart := offset
	if _, err := readOpaque(inner, &offset); err != nil {
		return nil, err
	}
	signatureKey, err := readOpaque(inner, &offset)
	if err != nil {
		return nil, err
	}
	credentialType := readUint16(inner, &offset)
	if credentialType != credentialTypeBasic {
		return nil, fmt.Errorf("unsupported key package credential type %d", credentialType)
	}
	identity, err := readOpaque(inner, &offset)
	if err != nil {
		return nil, err
	}
	if len(identity) != 8 {
		return nil, fmt.Errorf("expected 8-byte user identity, got %d bytes", len(identity))
	}

	versionsRaw, err := readOpaque(inner, &offset)
	if err != nil {
		return nil, err
	}
	suitesRaw, err := readOpaque(inner, &offset)
	if err != nil {
		return nil, err
	}
	if _, err := readOpaque(inner, &offset); err != nil {
		return nil, err
	}
	if _, err := readOpaque(inner, &offset); err != nil {
		return nil, err
	}
	credentialsRaw, err := readOpaque(inner, &offset)
	if err != nil {
		return nil, err
	}
	if !containsUint16(versionsRaw, protocolVersionMLS10) {
		return nil, fmt.Errorf("key package does not advertise MLS 1.0 support")
	}
	if !containsUint16(suitesRaw, daveCipherSuiteP256SHA256) {
		return nil, fmt.Errorf("key package does not advertise DAVE ciphersuite support")
	}
	if !containsUint16(credentialsRaw, credentialTypeBasic) {
		return nil, fmt.Errorf("key package does not advertise basic credential support")
	}

	if offset >= len(inner) {
		return nil, io.ErrUnexpectedEOF
	}
	source := inner[offset]
	offset++
	if source != 1 {
		return nil, fmt.Errorf("key package leaf node source must be key_package, got %d", source)
	}

	notBefore := readUint64(inner, &offset)
	notAfter := readUint64(inner, &offset)
	if notBefore != 0 || notAfter != math.MaxUint64 {
		return nil, fmt.Errorf("key package lifetime must be [0, 2^64-1], got [%d, %d]", notBefore, notAfter)
	}

	if _, err := readOpaque(inner, &offset); err != nil {
		return nil, err
	}
	leafSigFieldStart := offset
	leafSig, err := readOpaque(inner, &offset)
	if err != nil {
		return nil, err
	}
	leafNodeTBS := inner[leafNodeStart:leafSigFieldStart]

	if _, err := readOpaque(inner, &offset); err != nil {
		return nil, err
	}
	keyPackageSigFieldStart := offset
	if _, err := readOpaque(inner, &offset); err != nil {
		return nil, err
	}
	if offset != len(inner) {
		return nil, fmt.Errorf("key package contains trailing bytes")
	}

	leafKey, err := parseP256PublicKey(signatureKey)
	if err != nil {
		return nil, fmt.Errorf("invalid P-256 signature key: %w", err)
	}
	if !verifyWithLabel(leafKey, "LeafNodeTBS", leafNodeTBS, leafSig) {
		return nil, fmt.Errorf("leaf node signature verification failed")
	}
	keyPackageSigBytes, _, err := readOpaqueWithConsumed(inner, keyPackageSigFieldStart)
	if err != nil {
		return nil, err
	}
	if !verifyWithLabel(leafKey, "KeyPackageTBS", inner[:keyPackageSigFieldStart], keyPackageSigBytes) {
		return nil, fmt.Errorf("key package signature verification failed")
	}

	return &KeyPackage{
		Raw:          append([]byte(nil), raw...),
		Inner:        append([]byte(nil), inner...),
		UserID:       int64(binary.BigEndian.Uint64(identity)),
		Identity:     append([]byte(nil), identity...),
		SignatureKey: append([]byte(nil), signatureKey...),
	}, nil
}

func readOpaque(raw []byte, offset *int) ([]byte, error) {
	value, consumed, err := readOpaqueWithConsumed(raw, *offset)
	if err != nil {
		return nil, err
	}
	*offset += consumed
	return value, nil
}

func readOpaqueWithConsumed(raw []byte, offset int) ([]byte, int, error) {
	length, consumed, err := readVarint(raw, offset)
	if err != nil {
		return nil, 0, err
	}
	start := offset + consumed
	end := start + int(length)
	if end > len(raw) {
		return nil, 0, io.ErrUnexpectedEOF
	}
	return raw[start:end], consumed + int(length), nil
}

func readVarint(raw []byte, offset int) (uint32, int, error) {
	if offset >= len(raw) {
		return 0, 0, io.ErrUnexpectedEOF
	}
	b := raw[offset]
	prefix := b >> 6
	if prefix == 3 {
		return 0, 0, errors.New("invalid MLS varint prefix")
	}
	width := 1 << prefix
	if offset+width > len(raw) {
		return 0, 0, io.ErrUnexpectedEOF
	}
	value := uint32(b & 0x3f)
	for i := 1; i < width; i++ {
		value = (value << 8) | uint32(raw[offset+i])
	}
	return value, width, nil
}

func writeVarint(dst []byte, n int) []byte {
	switch {
	case n < 1<<6:
		return append(dst, byte(n))
	case n < 1<<14:
		v := uint16(0x4000 | n)
		return binary.BigEndian.AppendUint16(dst, v)
	case n < 1<<30:
		v := uint32(0x80000000 | n)
		return binary.BigEndian.AppendUint32(dst, v)
	default:
		panic("mls varint exceeds 30 bits")
	}
}

func appendOpaque(dst []byte, value []byte) []byte {
	dst = writeVarint(dst, len(value))
	return append(dst, value...)
}

func readUint16(raw []byte, offset *int) uint16 {
	if *offset+2 > len(raw) {
		panic(io.ErrUnexpectedEOF)
	}
	value := binary.BigEndian.Uint16(raw[*offset : *offset+2])
	*offset += 2
	return value
}

func readUint64(raw []byte, offset *int) uint64 {
	if *offset+8 > len(raw) {
		panic(io.ErrUnexpectedEOF)
	}
	value := binary.BigEndian.Uint64(raw[*offset : *offset+8])
	*offset += 8
	return value
}

func containsUint16(raw []byte, want uint16) bool {
	for i := 0; i+1 < len(raw); i += 2 {
		if binary.BigEndian.Uint16(raw[i:i+2]) == want {
			return true
		}
	}
	return false
}

func parseP256PublicKey(raw []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(elliptic.P256(), raw)
	if x == nil || y == nil {
		return nil, fmt.Errorf("failed to decode elliptic public key")
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

func signWithLabel(priv *ecdsa.PrivateKey, label string, content []byte) ([]byte, error) {
	toSign := marshalSignContent(label, content)
	digest := sha256.Sum256(toSign)
	return ecdsa.SignASN1(rand.Reader, priv, digest[:])
}

func verifyWithLabel(pub *ecdsa.PublicKey, label string, content, signature []byte) bool {
	toVerify := marshalSignContent(label, content)
	digest := sha256.Sum256(toVerify)
	return ecdsa.VerifyASN1(pub, digest[:], signature)
}

func marshalSignContent(label string, content []byte) []byte {
	out := appendOpaque(nil, []byte("MLS 1.0 "+label))
	return appendOpaque(out, content)
}

func refHash(label string, value []byte) ([]byte, error) {
	input := appendOpaque(nil, []byte(label))
	input = appendOpaque(input, value)
	h := crypto.SHA256.New()
	if _, err := h.Write(input); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
