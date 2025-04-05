package webhook

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func extractAttachmentID(s3 *S3Element) (attachmentID, channelId int64, err error) {
	rawoid, ook := s3.Object.UserMetadata["X-Amz-Meta-ObjectID"]
	rawcid, cok := s3.Object.UserMetadata["X-Amz-Meta-ChannelID"]
	if !ook || !cok {
		decodedValue, urlErr := url.QueryUnescape(s3.Object.Key)
		if urlErr != nil {
			err = fmt.Errorf("failed to decode object key: %w", urlErr)
			return
		}
		segments := strings.Split(decodedValue, "/")
		if len(segments) < 3 {
			err = errors.New("unable to parse key segments (bucket/channel/object) from key: " + decodedValue)
			return
		}
		rawcid = segments[1]
		rawoid = segments[2]
	}

	channelId, err = strconv.ParseInt(rawcid, 10, 64)
	if err != nil {
		err = fmt.Errorf("failed to parse channel ID '%s': %w", rawcid, err)
		return
	}

	attachmentID, err = strconv.ParseInt(rawoid, 10, 64)
	if err != nil {
		err = fmt.Errorf("failed to parse attachment ID '%s': %w", rawoid, err)
		return
	}

	return
}
