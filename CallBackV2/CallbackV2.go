package CallBackV2

import (
	"github.com/google/uuid"
	"github.com/ijnkawakaze/telegram-bot-api/CallBackV2/internal/callback"
	tgbotapi "github.com/ijnkawakaze/telegram-bot-api/Types"
	"time"
)

type Maker = callback.CallBackV2Maker
type StaticFunction = callback.StaticCallbackFunction
type LegacyStaticFunction = callback.LegacyStaticCallbackFunction
type Function = callback.CallbackFunction
type CleanUpFunction = callback.CleanUpFunction

func Init(logFunc func(err error)) error {
	return callback.InitCallBackV2(logFunc)
}
func Handler(update tgbotapi.Update) error {
	return callback.CallBackHandler(update)
}

func Register(maker Maker) (uuid.UUID, tgbotapi.InlineKeyboardMarkup, error) {
	return callback.RegisterCallback(maker)
}
func RegisterCustomTimeOut(maker Maker, timeOutDuration time.Duration) (uuid.UUID, tgbotapi.InlineKeyboardMarkup, error) {
	return callback.RegisterCallbackCustomTimeOut(maker, timeOutDuration)
}

func RegisterStatic(prefix string, callbackFunction StaticFunction) error {
	return callback.RegisterStaticCallback(prefix, callbackFunction.ToLegacyFunction())
}
func RegisterStaticLegacy(prefix string, callbackFunction callback.LegacyStaticCallbackFunction) error {
	return callback.RegisterStaticCallback(prefix, callbackFunction)
}
