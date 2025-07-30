package callback

import (
	"errors"
	"github.com/google/uuid"
	"github.com/ijnkawakaze/telegram-bot-api/CallBackV2/internal/Timer"
	"github.com/ijnkawakaze/telegram-bot-api/CallBackV2/internal/utils"
	tgbotapi "github.com/ijnkawakaze/telegram-bot-api/Types"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mutex sync.Mutex
var callbackV2Enabled bool = false
var timer Timer.Timer
var uuidMapper = make(map[uuid.UUID]*callbackV2)
var staticMapper = make(map[string]LegacyStaticCallbackFunction)

const DEFAULT_TIMEOUT = time.Minute
const UUID_LENGTH = 36

var EMPTY_MARKUP = tgbotapi.InlineKeyboardMarkup{make([][]tgbotapi.InlineKeyboardButton, 0)}

func InitCallBackV2(logFunc func(err error)) error {
	mutex.Lock()
	defer mutex.Unlock()
	if !callbackV2Enabled {
		callbackV2Enabled = true
		return realInitCallbackV2(logFunc)
	}
	if callbackV2Enabled {
		return utils.AlreadyEnabledError("CallbackV2")
	}
	panic("Code reach unreachable point")
}
func realInitCallbackV2(logFunc func(err error)) error {
	if logFunc == nil {
		timer = Timer.NewTimer()
	} else {
		timer = Timer.NewTimerWithLogFunc(logFunc)
	}
	return nil
}
func warpOnCleanUp(cleanUpFunc CleanUpFunction, v2 *callbackV2, uuid2 uuid.UUID) func() error {
	return func() error {
		mutex.Lock()
		defer mutex.Unlock()
		delete(uuidMapper, uuid2)
		return cleanUpFunc(nil, v2)
	}

}
func RegisterCallback(selection CallBackV2Maker) (uuid.UUID, tgbotapi.InlineKeyboardMarkup, error) {
	return RegisterCallbackCustomTimeOut(selection, DEFAULT_TIMEOUT)
}

func RegisterStaticCallback(prefix string, callbackFunction LegacyStaticCallbackFunction) error {
	if len(prefix) >= UUID_LENGTH {
		return errors.New("prefix too long")
	} else if utils.MapContain(staticMapper, prefix) {
		return errors.New("prefix has already been registered")
	} else {
		staticMapper[prefix] = callbackFunction
		return nil
	}
}
func RegisterCallbackCustomTimeOut(selection CallBackV2Maker, timeOutDuration time.Duration) (uuid.UUID, tgbotapi.InlineKeyboardMarkup, error) {
	mutex.Lock()
	defer mutex.Unlock()
	callbackUUID := timer.TimerGetAvailableUUID()
	markUP, callback := selection.toInlineKeyBoardMarkUp(callbackUUID)
	uuidMapper[callbackUUID] = &callback
	result := timer.TimerAddWithUUID(warpOnCleanUp(callback.onCleanUp, &callback, callbackUUID), timeOutDuration, callbackUUID)
	if result.IsErr() {
		return uuid.Nil, EMPTY_MARKUP, result.UnwarpErr()
	} else {
		return callbackUUID, markUP, nil
	}
}
func CallBackHandler(update tgbotapi.Update) error {
	query := strings.Split(update.CallbackQuery.Data, ",")
	if len(query) < 1 {
		return errors.New("invalid callback query : Empty CallbackQuery")
	}
	if len(query[0]) != UUID_LENGTH {
		//static query
		return handleStaticQuery(update, query[0])
	} else {
		// dynamic query
		if len(query) != 2 {
			return errors.New("invalid DynamicCallbackQuery: wrong length")
		}
		return handleDynamicQuery(update.CallbackQuery, query)
	}
}
func handleStaticQuery(query tgbotapi.Update, queryPrefix string) error {
	if !utils.MapContain(staticMapper, queryPrefix) {
		return errors.New("invalid Static CallbackQuery : prefix has not found")
	} else {
		return staticMapper[queryPrefix](query)
	}
}
func handleDynamicQuery(query *tgbotapi.CallbackQuery, queryS []string) error {
	targetUUID, UUIDErr := uuid.Parse(queryS[0])
	index, indexParseErr := strconv.Atoi(queryS[1])
	if UUIDErr != nil {
		return errors.New("invalid DynamicCallbackQuery : UUID In query It can not be parsed")
	}
	if indexParseErr != nil {
		return errors.New("invalid DynamicCallbackQuery : invalid Integer In query")
	}
	mutex.Lock()
	defer mutex.Unlock()
	if !utils.MapContain(uuidMapper, targetUUID) {
		return errors.New("invalid DynamicCallbackQuery : callback UUID Not Found")
	}
	targetCallback := uuidMapper[targetUUID]
	if index >= len(targetCallback.callbackFunctions) {
		return errors.New("invalid DynamicCallbackQuery : callback Index Out of Range")
	}
	callbackFunction := targetCallback.callbackFunctions[index]
	if callbackFunction == nil {
		return errors.New("invalid DynamicCallbackQuery : callback Function Not Found")
	}
	done, err := callbackFunction(query, targetCallback)
	if done {
		timer.TimerRemove(targetUUID)
		err2 := targetCallback.onCleanUp(query, targetCallback)
		return errors.Join(err2, err)
	}
	return err

}
