package tgbotapi

import (
	"github.com/ijnkawakaze/telegram-bot-api/CallBackV2"
	"github.com/ijnkawakaze/telegram-bot-api/Log"
	"github.com/ijnkawakaze/telegram-bot-api/Types"
	"runtime/debug"
	"strings"
	"time"
)

var b *Types.BotAPI
var logger = *Log.GetLogger()
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
	api                     Types.BotAPI
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

type callbackFunction = func(update Types.Update) error
type inlineMatcher struct {
	prefix    string
	processor callbackFunction
}
type genericMatchProcessor struct {
	MatchFunc func(Types.Update) bool
	Processor callbackFunction
}

func (b *Bot) NewProcessor(match func(Types.Update) bool, processor func(update Types.Update) error) {
	b.matchProcessorSlice = append(b.matchProcessorSlice,
		&genericMatchProcessor{
			MatchFunc: match,
			Processor: processor,
		},
	)
}

func (b *Bot) NewMemberProcessor(processor func(update Types.Update) error) {
	b.NewProcessor(func(update Types.Update) bool {
		return update.Message != nil && len(update.Message.NewChatMembers) > 0
	}, processor)
}

func (b *Bot) LeftMemberProcessor(processor func(update Types.Update) error) {
	b.NewProcessor(func(update Types.Update) bool {
		return update.Message != nil && update.Message.LeftChatMember != nil
	}, processor)
}

func (b *Bot) NewCallBackProcessor(callBackType string, processor func(update Types.Update) error) {
	CallBackV2.RegisterStatic()
}

func (b *Bot) NewCommandProcessor(command string, processor func(update Types.Update) error) {

	b.addProcessor(command, processor, b.commandProcessor)
}

func (b *Bot) NewPrivateCommandProcessor(command string, processor func(update Types.Update) error) {

	b.addProcessor(command, processor, b.privateCommandProcessor)
}

func (b *Bot) NewWaitMessageProcessor(waitMessage string, processor func(update Types.Update) error) {

	b.addProcessor(waitMessage, processor, b.waitMsgProcess)
}

func (b *Bot) NewPhotoMessageProcessor(command string, processor func(update Types.Update) error) {

	b.addProcessor(command, processor, b.photoCommandProcess)
}

func (b *Bot) NewReplyMessageProcessor(command string, processor func(update Types.Update) error) {

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
		logger.Printf("command %s is already added overriding \n", command)
	}
	funcMap[command] = processor
}
func recoverWarp(function callbackFunction) callbackFunction {
	return func(msg Types.Update) error {
		defer func() {
			if r := recover(); r != nil {
				s := string(debug.Stack())
				logger.Printf("Recovered err=%v, stack=%s\n", r, s)
			}
		}()
		return function(msg)
	}
}
func (bot *Bot) selectFunction(msg Types.Update) (callbackFunction, string) {
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
		return CallBackV2.Handler, "  "
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

func MakeBotFromAPI(api Types.BotAPI) *Bot {
	var bot = &Bot{api: api}
	bot.InitMap()
	return bot
}

func (bot *Bot) Run() {
	_ = CallBackV2.Init(nil)
	u := Types.NewUpdate(0)
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
				logger.Println("用户", msg.SentFrom().FullName(), "执行了", command, "操作")
			}
			err := recoverWarp(process)(msg)
			if err != nil {
				logger.Println("Plugin Error", err.Error())
			}
		}
	}
}
