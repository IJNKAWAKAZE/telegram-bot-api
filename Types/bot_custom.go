package Types

func convertAdminRightToInt(chatMember ChatMember) uint16 {
	var result = uint16(0)
	if chatMember.CanManageChat {
		result |= AdminCanManageChat
	}
	if chatMember.CanDeleteMessages {
		result |= AdminCanDeleteMessages
	}
	if chatMember.CanManageVideoChats {
		result |= AdminCanManageVideoChats
	}
	if chatMember.CanRestrictMembers {
		result |= AdminCanRestrictMembers
	}
	if chatMember.CanPromoteMembers {
		result |= AdminCanPromoteMembers
	}
	if chatMember.CanChangeInfo {
		result |= AdminCanChangeInfo
	}
	if chatMember.CanInviteUsers {
		result |= AdminCanInviteUsers
	}
	if chatMember.CanPostMessages {
		result |= AdminCanPostMessages
	}
	if chatMember.CanEditMessages {
		result |= AdminCanEditMessages
	}
	if chatMember.CanPinMessages {
		result |= AdminCanPinMessages
	}
	return result
}

const (
	AdminIsAnonymous = 1 << iota
	AdminCanManageChat
	AdminCanDeleteMessages
	AdminCanManageVideoChats
	AdminCanRestrictMembers
	AdminCanPromoteMembers
	AdminCanChangeInfo
	AdminCanInviteUsers
	AdminCanPostMessages
	AdminCanEditMessages
	AdminCanPinMessages
)

const (
	NoMessagesPermission = "noMessagesPermission" // 无消息权限
	AllPermissions       = "allPermissions"       // 全部权限
)

// RestrictChatMember 修改用户权限
func (bot *BotAPI) RestrictChatMember(charId, userId int64, t string) (*APIResponse, error) {
	permissions := &ChatPermissions{}
	if t == NoMessagesPermission {
		permissions = &ChatPermissions{
			CanSendMessages: false,
		}
	} else if t == AllPermissions {
		permissions = &ChatPermissions{
			CanSendMessages:       true,
			CanSendMediaMessages:  true,
			CanSendPolls:          true,
			CanSendOtherMessages:  true,
			CanAddWebPagePreviews: true,
			CanInviteUsers:        true,
			CanChangeInfo:         true,
			CanPinMessages:        true,
		}
	}
	restrictChatMemberConfig := RestrictChatMemberConfig{
		Permissions: permissions,
		ChatMemberConfig: ChatMemberConfig{
			ChatID: charId,
			UserID: userId,
		},
	}
	return bot.Request(restrictChatMemberConfig)
}

// BanChatMember 封禁用户
func (bot *BotAPI) BanChatMember(chatId, userId int64) (*APIResponse, error) {
	banChatMemberConfig := BanChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			ChatID: chatId,
			UserID: userId,
		},
		RevokeMessages: true,
	}
	return bot.Request(banChatMemberConfig)
}

// UnbanChatMember 解封用户
func (bot *BotAPI) UnbanChatMember(chatId, userId int64) (*APIResponse, error) {
	unbanChatMemberConfig := UnbanChatMemberConfig{
		ChatMemberConfig: ChatMemberConfig{
			ChatID: chatId,
			UserID: userId,
		},
		OnlyIfBanned: true,
	}
	return bot.Request(unbanChatMemberConfig)
}

const (
	CREATOR = iota
	ADMINISTRATOR
	RESTRICTED
	LEFT
	MEMBER
	KICKED
	UNKNOWN
)

// GetChatMemberStatus 获取用户的群组状态
func (bot *BotAPI) GetChatMemberStatus(chatId, userId int64) int {
	config := GetChatMemberConfig{
		ChatConfigWithUser{ChatID: chatId, UserID: userId},
	}
	member, _ := bot.GetChatMember(config)
	switch member.Status {
	case "creator":
		return CREATOR
	case "administrator":
		return ADMINISTRATOR
	case "restricted":
		return RESTRICTED
	case "left":
		return LEFT
	case "member":
		return MEMBER
	case "kicked":
		return KICKED
	default:
		return UNKNOWN
	}
}

// IsAdmin 是否是创建者或管理员
func (bot *BotAPI) IsAdmin(chatId, userId int64) bool {
	return bot.IsAdminWithPermissions(chatId, userId, 0)
}

// IsAdminWithPermissions 权限检查
func (bot *BotAPI) IsAdminWithPermissions(chatId, userId int64, requiredPermissions uint16) bool {
	getChatMemberConfig := GetChatMemberConfig{
		ChatConfigWithUser: ChatConfigWithUser{
			ChatID: chatId,
			UserID: userId,
		},
	}
	memberInfo, _ := bot.GetChatMember(getChatMemberConfig)
	if memberInfo.Status == "creator" {
		return true
	} else if memberInfo.Status == "administrator" {
		if requiredPermissions == 0 {
			return true
		}
		adminPermission := convertAdminRightToInt(memberInfo)
		if adminPermission&requiredPermissions == requiredPermissions {
			return true
		}
	}
	return false
}
