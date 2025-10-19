package historycallback

import (
	"ai_tg_search/redidb"
	botkeyboards "ai_tg_search/tg_bot_utils/bot_keyboards"
	tgbsender "ai_tg_search/tg_bot_utils/tgb_sender"
	userstates "ai_tg_search/tg_bot_utils/user_states"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleHistoryCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	history_topic, rediErr := redidb.GetUserTopics(update.CallbackQuery.From.ID)
	if rediErr != nil {
		log.Println("Ошибка поиска истории пользователя: ", rediErr)
		tgbsender.SendCallbackMessage(bot, "Ошибка поиска истории пользователя", update, botkeyboards.MenuKeyboard)
		return
	}
	var historyKeyboard = tgbotapi.NewInlineKeyboardMarkup()
	for _, topic := range history_topic {
		historyKeyboard.InlineKeyboard = append(historyKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(topic, fmt.Sprintf("user-history-topic-%s", topic)),
		))
	}
	tgbsender.EditCallbackMessage(bot, "Выберите ниже топик из вашей истории", update, historyKeyboard)
}

func HandleHistoryTopicCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	callback_data := update.CallbackQuery.Data
	if strings.HasPrefix(callback_data, "user-history-topic") {
		topic := strings.TrimPrefix(callback_data, "user-history-topic-")
		news, redi_err := redidb.GetNewsByTopic(topic)
		if redi_err != nil || len(news.News) == 0 {
			log.Println("Ошибка в redis!\n", redi_err)
			tgbsender.SendCallbackMessage(bot, "Ошибка при поиске новостей!", update, botkeyboards.MenuKeyboard)
			return
		}
		userstates.UserNewsMu.Lock()
		userstates.UserHistory[update.CallbackQuery.From.ID] = &userstates.UserNewsSession{
			News:         news,
			CurrentIndex: 0,
		}
		userstates.UserNewsMu.Unlock()
		tgbsender.EditCallbackMessage(bot, "Выберите ниже топик из вашей истории", update, botkeyboards.HistoryOfNewsKeyboard[0])
	} else {
		tgbsender.SendCallbackMessage(bot, "Произошла ошибка поиска новостей! Попробуйте позже..", update, botkeyboards.MenuKeyboard)
		return
	}

}
