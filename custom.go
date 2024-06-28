package tgbotapi

import (
	"runtime/debug"
	"strings"
	"time"
)

var b *BotAPI

const (
	NoMessagesPermission = "noMessagesPermission" // 无消息权限
	AllPermissions       = "allPermissions"       // 全部权限
)

var WaitMessage = make(map[int64]interface{})

type Bot struct {
	matchProcessorSlice     []*genericMatchProcessor
	commandProcessor        map[string]callbackFunction
	privateCommandProcessor map[string]callbackFunction
	photoCommandProcess     map[string]callbackFunction
	replyCommandProcess     map[string]callbackFunction
	inlineQueryProcess      []*inlineMatcher
	callbackQueryProcess    map[string]callbackFunction
	waitMsgProcess          map[string]callbackFunction
}

var now = time.Now().Unix()

func (b *Bot) InitMap() {
	b.callbackQueryProcess = make(map[string]callbackFunction)
	b.commandProcessor = make(map[string]callbackFunction)
	b.privateCommandProcessor = make(map[string]callbackFunction)
	b.waitMsgProcess = make(map[string]callbackFunction)
	b.photoCommandProcess = make(map[string]callbackFunction)
	b.replyCommandProcess = make(map[string]callbackFunction)
}

type callbackFunction = func(update Update) error
type inlineMatcher struct {
	prefix    string
	processor callbackFunction
}
type genericMatchProcessor struct {
	MatchFunc func(Update) bool
	Processor callbackFunction
}

func (b *Bot) NewProcessor(match func(Update) bool, processor func(update Update) error) {
	b.matchProcessorSlice = append(b.matchProcessorSlice,
		&genericMatchProcessor{
			MatchFunc: match,
			Processor: processor,
		},
	)
}

func (b *Bot) NewMemberProcessor(processor func(update Update) error) {
	b.NewProcessor(func(update Update) bool {
		return update.Message != nil && len(update.Message.NewChatMembers) > 0
	}, processor)
}

func (b *Bot) LeftMemberProcessor(processor func(update Update) error) {
	b.NewProcessor(func(update Update) bool {
		return update.Message != nil && update.Message.LeftChatMember != nil
	}, processor)
}

func (b *Bot) NewCallBackProcessor(callBackType string, processor func(update Update) error) {

	b.addProcessor(callBackType, processor, b.callbackQueryProcess)
}

func (b *Bot) NewCommandProcessor(command string, processor func(update Update) error) {

	b.addProcessor(command, processor, b.commandProcessor)
}

func (b *Bot) NewPrivateCommandProcessor(command string, processor func(update Update) error) {

	b.addProcessor(command, processor, b.privateCommandProcessor)
}

func (b *Bot) NewWaitMessageProcessor(waitMessage string, processor func(update Update) error) {

	b.addProcessor(waitMessage, processor, b.waitMsgProcess)
}

func (b *Bot) NewPhotoMessageProcessor(command string, processor func(update Update) error) {

	b.addProcessor(command, processor, b.photoCommandProcess)
}

func (b *Bot) NewReplyMessageProcessor(command string, processor func(update Update) error) {

	b.addProcessor(command, processor, b.replyCommandProcess)
}

func (b *Bot) NewInlineQueryProcessor(command string, processor callbackFunction) {
	b.inlineQueryProcess = append(b.inlineQueryProcess, &inlineMatcher{
		prefix:    command,
		processor: processor,
	})
}
func (b *Bot) addProcessor(command string, processor callbackFunction, funcMap map[string]callbackFunction) {
	_, ok := funcMap[command]
	if ok {
		log.Printf("command %s is already added overriding \n", command)
	}
	funcMap[command] = processor
}
func recoverWarp(function callbackFunction) callbackFunction {
	return func(msg Update) error {
		defer func() {
			if r := recover(); r != nil {
				s := string(debug.Stack())
				log.Printf("Recovered err=%v, stack=%s\n", r, s)
			}
		}()
		return function(msg)
	}
}
func (bot *Bot) selectFunction(msg Update) (callbackFunction, string) {
	// generic first
	for _, k := range bot.matchProcessorSlice {
		if k.MatchFunc(msg) {
			return k.Processor, ""
		}
	}
	if msg.Message != nil {
		//photo related cmd
		if len(msg.Message.Photo) > 0 {
			me, _ := b.GetMe()
			suffix := "@" + me.UserName
			command, _ := strings.CutSuffix(msg.Message.Caption, suffix)
			command = strings.Split(command, " ")[0]
			result, ok := bot.photoCommandProcess[command]
			if ok {
				return result, command
			}
		}
		if msg.Message.ReplyToMessage != nil {
			me, _ := b.GetMe()
			suffix := "@" + me.UserName
			command, _ := strings.CutSuffix(msg.Message.Text, suffix)
			command = strings.Split(command, " ")[0]
			result, ok := bot.replyCommandProcess[command]
			if ok {
				return result, command
			}
		}
		//private cmd
		command := msg.Message.Command()
		if msg.Message.Chat.IsPrivate() {
			result, ok := bot.privateCommandProcessor[command]
			if ok {
				return result, command
			}
			res, ok := WaitMessage[msg.Message.From.ID]
			if ok {
				waitMsg, is_str := res.(string)
				if is_str {
					return bot.waitMsgProcess[waitMsg], waitMsg
				}
			}
		}
		//normal command
		result, ok := bot.commandProcessor[command]
		if ok {
			return result, command
		}
	}
	// callback
	if msg.CallbackQuery != nil {
		callback_q := strings.Split(msg.CallbackData(), ",")[0]
		result, ok := bot.callbackQueryProcess[callback_q]
		if ok {
			return result, ""
		}
	}
	//inline Q
	if msg.InlineQuery != nil {
		for _, v := range bot.inlineQueryProcess {
			if strings.HasPrefix(msg.InlineQuery.Query, v.prefix) {
				return v.processor, ""
			}
		}
	}
	return nil, ""
}

func (_ *BotAPI) AddHandle() *Bot {
	b := &Bot{}
	b.InitMap()
	return b
}

func (bot *Bot) Run() {
	u := NewUpdate(0)
	u.Timeout = 60
	updates := b.GetUpdatesChan(u)
	if updates == nil {
		panic("updates is nil")
	}
	for {
		msg := <-updates
		if msg.Message != nil && msg.Message.Time().Unix() < now {
			continue
		}
		if msg.Message != nil && msg.Message.IsCommand() && msg.Message.From.ID == 136817688 {
			msg.Message.Delete()
			continue
		}

		process, command := bot.selectFunction(msg)
		if process != nil {
			if command != "" {
				log.Println("用户", msg.SentFrom().FullName(), "执行了", command, "操作")
			}
			err := recoverWarp(process)(msg)
			if err != nil {
				log.Println("Plugin Error", err.Error())
			}
		}
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
