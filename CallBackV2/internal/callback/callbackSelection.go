package callback

import (
	"github.com/google/uuid"
	tgbotapi "github.com/ijnkawakaze/telegram-bot-api/Types"
	"strconv"
	"strings"
)

var builder strings.Builder

type CleanUpFunction func(*tgbotapi.CallbackQuery, *callbackV2) error
type CallbackFunction = func(*tgbotapi.CallbackQuery, *callbackV2) (bool, error)
type StaticCallbackFunction func(*tgbotapi.CallbackQuery) error
type LegacyStaticCallbackFunction = func(update tgbotapi.Update) error
type callbackSelection struct {
	tgbotapi.InlineKeyboardButton
	callbackFunction CallbackFunction
}

func (c callbackSelection) toInlineKeyBoardButton(uuid uuid.UUID, index int) (tgbotapi.InlineKeyboardButton, CallbackFunction) {
	builder.Reset()
	var callbackString string
	if c.callbackFunction != nil {
		builder.WriteString(uuid.String())
		builder.WriteByte(',')
		builder.WriteString(strconv.Itoa(index))
		callbackString = builder.String()
		c.CallbackData = &callbackString
	}
	return c.InlineKeyboardButton, c.callbackFunction
}
func (f StaticCallbackFunction) ToLegacyFunction() LegacyStaticCallbackFunction {
	return func(update tgbotapi.Update) error {
		return f(update.CallbackQuery)
	}
}
