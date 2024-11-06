package dto

type AttachmentUpload struct {
	Id        int64  `json:"id"`
	ChannelId int64  `json:"channel_id"`
	FileName  string `json:"file_name"`
	UploadURL string `json:"upload_url"`
}
