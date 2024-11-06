package permissions

type RolePermission int64

const (
	PermissionViewChannel RolePermission = 1 << iota
	PermissionSendMessages
	PermissionCreateChannels
	PermissionAddReactions
	PermissionRemoveReactions
	PermissionManageChannels
	PermissionManageGuild
	PermissionAdministrator
	PermissionKickMembers
	PermissionBanMembers
	PermissionManageMessages
	PermissionChangeNickname
	PermissionManageThreads
	PermissionCreateThreads
)
