package newscallback

import (
	"ai_tg_search/godb"
	perplexityapi "ai_tg_search/perplexity_api"
	"ai_tg_search/struct_types/newstypes"
	botkeyboards "ai_tg_search/tg_bot_utils/bot_keyboards"
	tgbsender "ai_tg_search/tg_bot_utils/tgb_sender"
	userstates "ai_tg_search/tg_bot_utils/user_states"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Обрабатывает callback запросы для поиска или истории новостей в зависимости от данных
// callback data = user-get-(history/news)
func HandleSearchOrHistoryCallback(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if update.CallbackQuery == nil {
		return
	}

	user_id := update.CallbackQuery.From.ID
	println(user_id)
	callback_data := update.CallbackQuery.Data

	//Проверка что за callback
	var user_state *userstates.UserNewsSession

	var keyboards [3]tgbotapi.InlineKeyboardMarkup

	if strings.Contains(callback_data, "search") {
		//Клавиатура
		keyboards = botkeyboards.SearchNewsKeyboard
		//Получение индекса
		userstates.UserNewsMu.RLock()
		user_state = userstates.UserSearch[user_id]
		userstates.UserNewsMu.RUnlock()
	} else if strings.Contains(callback_data, "history") {
		keyboards = botkeyboards.HistoryOfNewsKeyboard
		userstates.UserNewsMu.RLock()
		user_state = userstates.UserHistory[user_id]
		userstates.UserNewsMu.RUnlock()
	}

	// Проверка движения callback
	var callback_type string
	if strings.Contains(callback_data, "next") {
		callback_type = "next"
	} else if strings.Contains(callback_data, "prev") {
		callback_type = "prev"
	}

	// Получение индекса пользователя

	current_user_index := user_state.CurrentIndex

	switch callback_type {
	case "next":
		current_user_index++
	case "prev":
		current_user_index--
	}
	println(user_id)

	if strings.Contains(callback_data, "search") {
		userstates.UserNewsMu.Lock()
		userstates.UserSearch[user_id].CurrentIndex = current_user_index
		userstates.UserNewsMu.Unlock()
	} else if strings.Contains(callback_data, "history") {
		userstates.UserNewsMu.Lock()
		userstates.UserHistory[user_id].CurrentIndex = current_user_index
		userstates.UserNewsMu.Unlock()
	}

	text_to_send := createNewsToText(user_state.News, current_user_index)

	switch current_user_index {
	case 0:
		tgbsender.EditCallbackMessage(bot, text_to_send, *update, keyboards[0])
		return
	case user_state.News.TotalResults - 1:
		tgbsender.EditCallbackMessage(bot, text_to_send, *update, keyboards[2])
		return
	default:
		tgbsender.EditCallbackMessage(bot, text_to_send, *update, keyboards[1])
		return
	}

}

// Обслуживает команлу на /search
func HandleSearchCommand(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	user_id := update.Message.From.ID
	text := update.Message.CommandArguments()
	if text == "" {
		log.Println("ОШИБКА ТЕКСТ ПУСТОЙ")
		tgbsender.SendMessage(bot, "Введите топик для поиска новостей!", *update, botkeyboards.MenuKeyboard)
		return
	}
	requests_remaining, db_err := godb.DecrementUserRequests(user_id)
	if db_err != nil {
		log.Println("Ошибка дб при получении запрсоов пользователя: ", db_err)
		tgbsender.SendMessage(bot, "Кажется запросов больше нет!\n\nПополните запросы для продолжения работы с ботом", *update, botkeyboards.MenuKeyboard)
		return
	}

	tgbsender.SendMessage(bot, fmt.Sprintf("Начался поиск новостей!\nОставшееся количество запросов: %d", requests_remaining), *update, tgbotapi.NewInlineKeyboardMarkup())

	json_response, err := perplexityapi.FindNews(text, user_id)
	if err != nil {
		log.Println("Error searching news:", err)
		_, increment_err := godb.IncrementUserRequests(user_id)
		if increment_err != nil {
			log.Println("Ошибка возвращении запросов: ", increment_err)
			tgbsender.SendMessage(bot, "Ошибка при возвращении вашего запроса!\nОбратитесь в поддержку для его возвращения", *update, botkeyboards.MenuKeyboard)
			return
		}
		tgbsender.SendMessage(bot, "Ошибка при поиске новостей. Пожалуйста, попробуйте позже.", *update, botkeyboards.MenuKeyboard)
		return
	}
	//println(update.Message.From.ID)
	userstates.UserNewsMu.Lock()
	userstates.UserSearch[user_id] = &userstates.UserNewsSession{
		News:         json_response,
		CurrentIndex: 0,
	}
	userstates.UserNewsMu.Unlock()

	text_to_send := createNewsToText(json_response, 0)

	tgbsender.SendMessage(bot, text_to_send, *update, botkeyboards.SearchNewsKeyboard[0])
}

// Создает из NewsResponse готовый к отправке текст
func createNewsToText(news *newstypes.NewsResponse, index int) string {
	news_by_number := news.News[index]
	text := fmt.Sprintf("%s\n\n%s\n", news_by_number.Title, news_by_number.Summary)
	return text
}
