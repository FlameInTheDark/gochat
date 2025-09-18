package dto

type AttachmentUpload struct {
	Id        int64  `json:"id" example:"2230469276416868352"`         // Attachment ID
	ChannelId int64  `json:"channel_id" example:"2230469276416868352"` // Channel ID the attachment was sent to
	FileName  string `json:"file_name" example:"image.png"`            // File name
	UploadURL string `json:"upload_url"`                               // Upload URL. S3 presigned URL
}
