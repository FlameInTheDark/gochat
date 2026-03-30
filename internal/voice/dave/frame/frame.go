package frame

import "bytes"

const (
	MagicMarkerHigh          = 0xFA
	MagicMarkerLow           = 0xFA
	MagicMarker              = uint16(0xFAFA)
	MinSupplementalDataBytes = 11
)

var silencePacket = []byte{0xF8, 0xFF, 0xFE}

func HasMagicMarker(payload []byte) bool {
	if len(payload) < 3 {
		return false
	}
	return payload[len(payload)-2] == MagicMarkerHigh && payload[len(payload)-1] == MagicMarkerLow
}

func LooksLikeProtocolFrame(payload []byte) bool {
	if len(payload) < MinSupplementalDataBytes {
		return false
	}
	if !HasMagicMarker(payload) {
		return false
	}
	supplementalSize := int(payload[len(payload)-3])
	if supplementalSize < MinSupplementalDataBytes {
		return false
	}
	if supplementalSize > len(payload) {
		return false
	}
	return true
}

func ShouldDropForReceiver(payload []byte, receiverSupportsDAVE bool, transitionInFlight bool) bool {
	return transitionInFlight && !receiverSupportsDAVE && LooksLikeProtocolFrame(payload)
}

func IsSilencePacket(payload []byte) bool {
	return bytes.Equal(payload, silencePacket)
}
