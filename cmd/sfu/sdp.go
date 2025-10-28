package main

import (
	"strconv"
	"strings"

	"github.com/pion/sdp/v3"
)

// limitAudioBitrateInSDP injects bandwidth limits and OPUS fmtp maxaveragebitrate
// into all audio media sections. If parsing fails, returns the original SDP.
// maxBps must be > 0.
func limitAudioBitrateInSDP(sdpIn string, maxBps uint64) string {
	if maxBps == 0 {
		return sdpIn
	}

	var desc sdp.SessionDescription
	if err := desc.UnmarshalString(sdpIn); err != nil {
		return sdpIn
	}

	// Iterate all media descriptions and apply to audio
	for _, md := range desc.MediaDescriptions {
		if md == nil || !strings.EqualFold(md.MediaName.Media, "audio") {
			continue
		}

		// Update/insert bandwidth lines (TIAS = bits per second, AS = kilobits per second)
		kbps := maxBps / 1000
		var hasTIAS, hasAS bool
		for i := range md.Bandwidth {
			switch strings.ToUpper(md.Bandwidth[i].Type) {
			case "TIAS":
				md.Bandwidth[i].Bandwidth = maxBps
				hasTIAS = true
			case "AS":
				md.Bandwidth[i].Bandwidth = kbps
				hasAS = true
			}
		}
		if !hasTIAS {
			md.Bandwidth = append(md.Bandwidth, sdp.Bandwidth{Type: "TIAS", Bandwidth: maxBps})
		}
		if !hasAS {
			md.Bandwidth = append(md.Bandwidth, sdp.Bandwidth{Type: "AS", Bandwidth: kbps})
		}

		// Discover Opus payload type from rtpmap
		opusPT := ""
		for _, a := range md.Attributes {
			if strings.EqualFold(a.Key, "rtpmap") {
				// format: "<pt> codec/<clock>[/channels]"
				parts := strings.Fields(a.Value)
				if len(parts) >= 2 && strings.Contains(strings.ToLower(parts[1]), "opus/") {
					opusPT = parts[0]
					break
				}
			}
		}
		if opusPT == "" {
			continue
		}

		// Clamp to allowed range per RFC7587 (6000..510000)
		if maxBps < 6000 {
			maxBps = 6000
		} else if maxBps > 510000 {
			maxBps = 510000
		}
		maxStr := strconv.FormatUint(maxBps, 10)

		// Update or add fmtp for opus payload
		updated := false
		for i := range md.Attributes {
			a := &md.Attributes[i]
			if !strings.EqualFold(a.Key, "fmtp") {
				continue
			}
			// a.Value like: "111 minptime=10;useinbandfec=1"
			if !strings.HasPrefix(strings.TrimSpace(a.Value), opusPT+" ") && !strings.EqualFold(strings.TrimSpace(a.Value), opusPT) {
				continue
			}

			// split into "<pt>" and params
			rest := strings.TrimSpace(strings.TrimPrefix(a.Value, opusPT))
			rest = strings.TrimSpace(rest)
			params := rest
			if params == "" {
				a.Value = opusPT + " maxaveragebitrate=" + maxStr
				updated = true
				break
			}
			// modify/append maxaveragebitrate
			kvs := strings.Split(params, ";")
			found := false
			for j := range kvs {
				kv := strings.TrimSpace(kvs[j])
				if kv == "" {
					continue
				}
				if strings.HasPrefix(strings.ToLower(kv), "maxaveragebitrate=") {
					kvs[j] = "maxaveragebitrate=" + maxStr
					found = true
					break
				}
			}
			if !found {
				kvs = append(kvs, "maxaveragebitrate="+maxStr)
			}
			// rebuild
			a.Value = opusPT + " " + strings.Join(kvs, ";")
			updated = true
			break
		}
		if !updated {
			// No existing fmtp for opus, add one
			md.Attributes = append(md.Attributes, sdp.NewAttribute("fmtp", opusPT+" maxaveragebitrate="+maxStr))
		}
	}

	if out, err := desc.Marshal(); err == nil {
		return string(out)
	}
	return sdpIn
}
