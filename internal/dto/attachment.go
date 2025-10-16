package dto

type AttachmentUpload struct {
	Id        int64  `json:"id" example:"2230469276416868352"`         // Attachment ID
	ChannelId int64  `json:"channel_id" example:"2230469276416868352"` // Channel ID the attachment was sent to
	FileName  string `json:"file_name" example:"image.png"`            // File name
	// UploadURL removed: uploads are now handled by POST /api/v1/attachments/{channel_id}/{attachment_id}
}
