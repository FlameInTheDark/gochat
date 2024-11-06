package webhook

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

func extractAttachmentID(s3 *S3Element) (attachmentID, channelId int64, err error) {
	rawoid, ook := s3.Object.UserMetadata["X-Amz-Meta-ObjectID"]
	rawcid, cok := s3.Object.UserMetadata["X-Amz-Meta-ChannelID"]
	if !ook || !cok {
		decodedValue, err := url.QueryUnescape(s3.Object.Key)
		if err != nil {
			return 0, 0, err
		}
		segments := strings.Split(decodedValue, "/")
		if len(segments) < 2 {
			return 0, 0, errors.New("unable to parse key to ids")
		}
		rawcid = segments[0]
		rawoid = segments[1]
	}
	channelId, err = strconv.ParseInt(rawcid, 10, 64)
	attachmentID, err = strconv.ParseInt(rawoid, 10, 64)
	return
}
