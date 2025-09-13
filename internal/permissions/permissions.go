package permissions

type RolePermission int64

const (
	PermServerViewChannels RolePermission = 1 << iota
	PermServerManageChannels
	PermServerManageRoles
	PermServerViewAuditLog
	PermServerManage
	PermMembershipCreateInvite
	PermMembershipChangeNickname
	PermMembershipManageNickname
	PermMembershipKickMembers
	PermMembershipBanMembers
	PermMembershipTimeoutMembers
	PermTextSendMessage
	PermTextSendMessageInThreads
	PermTextCreateThreads
	PermTextAttachFiles
	PermTextAddReactions
	PermTextMentionRoles
	PermTextManageMessages
	PermTextManageThreads
	PermTextReadMessageHistory
	PermVoiceConnect
	PermVoiceSpeak
	PermVoiceVideo
	PermVoiceMuteMembers
	PermVoiceDeafenMembers
	PermVoiceMoveMembers
	PermAdministrator
)

var DefaultPermissions = CreatePermissions(
	PermServerViewChannels,
	PermMembershipCreateInvite,
	PermMembershipChangeNickname,
	PermTextSendMessage,
	PermTextSendMessageInThreads,
	PermTextCreateThreads,
	PermTextAddReactions,
	PermTextAttachFiles,
	PermTextAddReactions,
	PermTextReadMessageHistory,
	PermVoiceConnect,
	PermVoiceSpeak,
	PermVoiceVideo)
