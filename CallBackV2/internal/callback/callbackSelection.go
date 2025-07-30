package callback

import (
	"github.com/google/uuid"
	Type "github.com/ijnkawakaze/telegram-bot-api/Types"
	"strconv"
	"strings"
)

var builder strings.Builder

type CleanUpFunction func(*Type.CallbackQuery, *callbackV2) error
type CallbackFunction = func(*Type.CallbackQuery, *callbackV2) (bool, error)
type StaticCallbackFunction func(*Type.CallbackQuery) error
type LegacyStaticCallbackFunction = func(update Type.Update) error
type callbackSelection struct {
	Type.InlineKeyboardButton
	callbackFunction CallbackFunction
}

func (c callbackSelection) toInlineKeyBoardButton(uuid uuid.UUID, index int) (Type.InlineKeyboardButton, CallbackFunction) {
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
	return func(update Type.Update) error {
		return f(update.CallbackQuery)
	}
}
