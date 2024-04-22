package tgbotapi

var b *BotAPI

const (
	NoMessagesPermission = "noMessagesPermission" // 无消息权限
	AllPermissions       = "allPermissions"       // 全部权限
)

// IsAdmin 是否是创建者或管理员
func (bot *BotAPI) IsAdmin(chatId, userId int64) bool {
	getChatMemberConfig := GetChatMemberConfig{
		ChatConfigWithUser: ChatConfigWithUser{
			ChatID: chatId,
			UserID: userId,
		},
	}
	memberInfo, _ := bot.GetChatMember(getChatMemberConfig)
	if memberInfo.Status != "creator" && memberInfo.Status != "administrator" {
		return false
	}
	return true
}

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

// FullName 获取用户全名
func (u *User) FullName() string {
	if u == nil {
		return ""
	}

	name := u.FirstName
	if u.LastName != "" {
		name += " " + u.LastName
	}

	return name
}

// Delete 删除消息
func (m *Message) Delete() (*APIResponse, error) {
	delMsg := NewDeleteMessage(m.Chat.ID, m.MessageID)
	return b.Request(delMsg)
}

// Delete 删除回调消息
func (c *CallbackQuery) Delete() (*APIResponse, error) {
	delMsg := NewDeleteMessage(c.Message.Chat.ID, c.Message.MessageID)
	return b.Request(delMsg)
}

// Answer 回调响应
func (c *CallbackQuery) Answer(showAlert bool, text string) (*APIResponse, error) {
	answer := NewCallback(c.ID, text)
	if showAlert {
		answer = NewCallbackWithAlert(c.ID, text)
	}
	return b.Request(answer)
}
