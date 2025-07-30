package Types

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
func (m *Message) Delete(api BotAPI) (*APIResponse, error) {
	delMsg := NewDeleteMessage(m.Chat.ID, m.MessageID)
	return api.Request(delMsg)
}

// Delete 删除回调消息
func (c *CallbackQuery) Delete(api BotAPI) (*APIResponse, error) {
	delMsg := NewDeleteMessage(c.Message.Chat.ID, c.Message.MessageID)
	return api.Request(delMsg)
}

// Answer 回调响应
func (c *CallbackQuery) Answer(showAlert bool, text string, api BotAPI) (*APIResponse, error) {
	answer := NewCallback(c.ID, text)
	if showAlert {
		answer = NewCallbackWithAlert(c.ID, text)
	}
	return api.Request(answer)
}
