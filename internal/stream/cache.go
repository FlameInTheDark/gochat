package stream

import "strconv"

func fmtInt64(v int64) string { return strconv.FormatInt(v, 10) }

func RouteKey(streamID int64) string {
	return "stream:route:" + fmtInt64(streamID)
}

func MetaKey(streamID int64) string {
	return "stream:meta:" + fmtInt64(streamID)
}

func UserKey(userID int64) string {
	return "stream:user:" + fmtInt64(userID)
}

func ChannelKey(channelID int64) string {
	return "stream:channel:" + fmtInt64(channelID)
}

func RebindKey(streamID int64) string {
	return "stream:rebind:" + fmtInt64(streamID)
}
