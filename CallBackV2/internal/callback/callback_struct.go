package callback

import (
	"github.com/google/uuid"
	tgbotapi "github.com/ijnkawakaze/telegram-bot-api"
	"github.com/ijnkawakaze/telegram-bot-api/CallBackV2/internal/utils"
	"sync"
)

type CallBackV2Maker struct {
	selections        [][]callbackSelection
	callbackFunctions []CallbackFunction
	globalData        map[string]any
	onCleanUp         CleanUpFunction
}

type callbackV2 struct {
	callbackUUID      uuid.UUID
	callbackFunctions []CallbackFunction
	globalData        map[string]any
	onCleanUp         CleanUpFunction
	mutexLock         sync.Mutex
}

func (c CallBackV2Maker) toInlineKeyBoardMarkUp(uuid2 uuid.UUID) (tgbotapi.InlineKeyboardMarkup, callbackV2) {
	funcIndex := 0
	callbackFunctions := make([]CallbackFunction, 0)
	inlineButtons := utils.FMap(c.selections, func(selections []callbackSelection) []tgbotapi.InlineKeyboardButton {
		return utils.FMap(selections, func(s callbackSelection) tgbotapi.InlineKeyboardButton {
			result, function := s.toInlineKeyBoardButton(uuid2, funcIndex)
			callbackFunctions = append(callbackFunctions, function)
			funcIndex++
			return result
		})
	})
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: inlineButtons}, callbackV2{
		callbackUUID:      uuid2,
		callbackFunctions: callbackFunctions,
		globalData:        c.globalData,
		onCleanUp:         c.onCleanUp,
		mutexLock:         sync.Mutex{},
	}

}
