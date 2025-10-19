package profilecallback

import (
	"ai_tg_search/godb"
	botkeyboards "ai_tg_search/tg_bot_utils/bot_keyboards"
	tgbsender "ai_tg_search/tg_bot_utils/tgb_sender"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func HandleProfileCallback(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	info_from := update.CallbackQuery.From

	user_info, db_err := godb.GetUser(info_from.ID)
	if db_err != nil {
		tgbsender.SendCallbackMessage(bot, "Ошибка поиска пользователя! Повторите попытку позже", update, botkeyboards.MenuKeyboard)
		return
	}

	text_to_send := fmt.Sprintf("👨‍🏫 Ваш профиль\nЮзернейм: %s\nКоличество запросов: %d\nВаша реферальная ссылка:\n%s", info_from.UserName, user_info.RequestsRemaining, "Попозже добавлю")

	tgbsender.EditCallbackMessage(bot, text_to_send, update, botkeyboards.ProfileKeyboard)
}
