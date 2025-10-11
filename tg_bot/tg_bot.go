package tgbot

import (
	godb "ai_tg_search/godb"
	perplexityapi "ai_tg_search/perplexity_api"
	"ai_tg_search/redidb"
	newstypes "ai_tg_search/struct_types/newstypes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const REFERRAL_LINK = "https://t.me/newsly_search_bot?start=ref_"

var bot *tgbotapi.BotAPI

var (
	UserState   = make(map[int64]string)
	UserStateMu sync.RWMutex
)

var (
	UserData   = make(map[int64]string)
	UserDataMu sync.RWMutex
)

var (
	UserHistory        = make(map[int64]*newstypes.NewsResponse)
	UserHistoryCurrent = make(map[int64]int)
	UserHistoryMu      sync.RWMutex
)

// Инлайн клавиатуры пользователя
//
// startKeyboard
var startKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Начать поиск", "search"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Помощь", "help"),
		tgbotapi.NewInlineKeyboardButtonData("Реферал", "refferal"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Профиль", "profile"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("Телеграмм канал", "https://t.me/newsly_search"),
	),
)

// menu keyboard
var menuKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
	),
)

// профиль колбек
var profileKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("История запросов", "history"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Помощь", "help"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
	),
)

var CallbackNews = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️", "user-get-history-prev"),
		tgbotapi.NewInlineKeyboardButtonData("➡️", "user-get-history-next"),
	),
)

var CallbackNewsFirst = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
	),
	tgbotapi.NewInlineKeyboardRow(
		//tgbotapi.NewInlineKeyboardButtonData("⬅️", "news-history-next"),
		tgbotapi.NewInlineKeyboardButtonData("➡️", "user-get-history-next"),
	),
)

var CallbackNewsLast = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➡️", "user-get-history-next"),
		//tgbotapi.NewInlineKeyboardButtonData("➡️", "news-history-prev"),
	),
)

func InitBot(token string) {
	bot_instant, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}
	bot = bot_instant
	bot.Debug = false //Дебаггиг бота

	updateConfig := tgbotapi.NewUpdate(0) // Инициализация конфигурации обновлений

	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)
	log.Printf("------------\nBot started his job with username %s\n------------", bot.Self.UserName)
	for update := range updates {
		go func(update tgbotapi.Update) {
			if update.Message != nil {
				if update.Message.IsCommand() {
					handleCommand(update)
				} else {
					handleMessage(update)
				}
			} else if update.CallbackQuery != nil {
				handleCallback(update)
			}
		}(update)
	}
}

// начало командной обработки
// Обработка команд пользователя
func handleCommand(update tgbotapi.Update) {
	switch update.Message.Command() {
	case "start":
		var refferal_string string
		args := update.Message.CommandArguments()
		if len(args) > 0 && strings.TrimPrefix(args, "ref_") != "" {
			refferal_string = strings.Split(args, "_")[1]
		}
		dbErr := godb.AddUserIfNotExists(update.Message.From.ID, refferal_string)
		if dbErr != nil {
			log.Println("Error adding user:", dbErr)
			sendMessage("Ошибка при добавлении пользователя. Пожалуйста, попробуйте позже.", update, tgbotapi.NewInlineKeyboardMarkup())
		}
		sendMessage("Добро пожаловать в newsly!\n\nБыстрый и удобный бот для поиска новостей на любую тематику.\n\nВыберите опцию ниже", update, startKeyboard)

	case "help":
		sendMessage("Для поиска новостей введите команду /search и тему, например: /search IT", update, menuKeyboard)

	case "search":
		requests_remaining, dbErr := godb.DecrementUserRequests(update.Message.From.ID)
		if dbErr != nil {
			log.Println(dbErr)
			sendMessage("Ошибка!\nСкорее всего количество запросов исчерпано", update, tgbotapi.NewInlineKeyboardMarkup())
		} else {
			text := update.Message.CommandArguments()
			sendMessage(fmt.Sprintf("Начался поиск новостей\nОставшееся количество запросов:%d\n  Ожидайте...", requests_remaining), update, tgbotapi.NewInlineKeyboardMarkup())
			json_response, err := perplexityapi.FindNews(text, update.Message.From.ID)
			if err != nil {
				log.Println("Error searching news:", err)
				sendMessage("Ошибка при поиске новостей. Пожалуйста, попробуйте позже.", update, menuKeyboard)
			}
			for _, news := range json_response.News {
				sendMessage(fmt.Sprintf("Заголовок: %s\n\n%s\n\nURL: %s", news.Title, news.Summary, news.URL), update, tgbotapi.NewInlineKeyboardMarkup())
			}
		}

	case "add":
		text := strings.Split(update.Message.CommandArguments(), " ")
		if len(text) > 1 {
			//Парсинг аргументов
			user_id := text[0]
			amount := text[1]
			//перевод user_id в int
			user_int_id, IDerr := strconv.Atoi(user_id)
			if IDerr != nil {
				sendMessage("Введите полноценный ID пользователя", update, tgbotapi.NewInlineKeyboardMarkup())
			}
			// int -> int64
			user_int64_id := int64(user_int_id)
			// перевод amount в int64
			amount_int, amountErr := strconv.Atoi(amount)
			if amountErr != nil {
				sendMessage("Введите полноценную сумму", update, tgbotapi.NewInlineKeyboardMarkup())
			}
			dbErr := godb.AddRequests(update.Message.From.ID, user_int64_id, amount_int)
			if dbErr != nil {
				log.Println(dbErr)
				sendMessage("Ошибка при добавлении запроса", update, tgbotapi.NewInlineKeyboardMarkup())
			}
			sendMessage("Запрос успешно добавлен", update, tgbotapi.NewInlineKeyboardMarkup())
		}
	default:

	}
}

func handleMessage(update tgbotapi.Update) {
	state := UserState[update.Message.Chat.ID]
	switch state {
	case "user-search":
		UserDataMu.Lock()
		UserData[update.Message.Chat.ID] = update.Message.Text
		UserDataMu.Unlock()
		UserStateMu.Lock()
		UserState[update.Message.Chat.ID] = "waiting-for-response"
		UserStateMu.Unlock()
		sendMessage("Запрос успешно сохранен. Ожидайте ответа от бота.", update, tgbotapi.NewInlineKeyboardMarkup())

	default:
		sendMessage("Неизвестная команда", update, tgbotapi.NewInlineKeyboardMarkup())
	}
}

func handleCallback(update tgbotapi.Update) {
	switch update.CallbackQuery.Data {
	case "menu":
		editMessage("Добро пожаловать в newsly!\n\nБыстрый и удобный бот для поиска новостей на любую тематику.\n\nВыберите опцию ниже", update, startKeyboard)
	case "search":
		UserStateMu.Lock()
		UserState[update.CallbackQuery.From.ID] = "user-search"
		UserStateMu.Unlock()
		editMessage("Введите запрос для поиска новостей\n\nЗапросом должен являться текст, по которому будет выполнен поиск новостей\nПример: Рынок криптовалют, лучшие тенденции, восходящие альт коины", update, tgbotapi.NewInlineKeyboardMarkup())
	case "profile":
		user, dbErr := godb.GetUser(update.CallbackQuery.From.ID)
		if dbErr != nil {
			log.Println("Ошибка поиска пользователя: ", dbErr)
			editMessage("Ошибка поиска пользователя", update, menuKeyboard)
			return
		}
		editMessage(fmt.Sprintf("Ваш профиль\n\nНикнейм: %s\nКол-во запросов: %d\n", update.CallbackQuery.From.UserName, user.RequestsRemaining), update, profileKeyboard)
	case "history":
		history_topic, rediErr := redidb.GetUserTopics(update.CallbackQuery.From.ID)
		if rediErr != nil {
			log.Println("Ошибка поиска истории пользователя: ", rediErr)
			editMessage("Ошибка поиска истории пользователя", update, profileKeyboard)
			return
		}
		var historyKeyboard = tgbotapi.NewInlineKeyboardMarkup()
		for _, topic := range history_topic {
			historyKeyboard.InlineKeyboard = append(historyKeyboard.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(topic, fmt.Sprintf("user-history-topic-%s", topic)),
			))
		}
		editMessage("Выберите ниже топик из вашей истории", update, historyKeyboard)
	case "help":
		sendMessage("Добро пожаловать в наш бот!\nОн создан специально для быстрого поиска новостей.\n\nНажмите ниже на inline кнопку", update, tgbotapi.NewInlineKeyboardMarkup())
	case "refferal":
		ref_uuid, dbErr := godb.GetUserReferralLink(update.CallbackQuery.From.ID)
		if dbErr != nil {
			log.Println("Ошибка поиска реферальной ссылки: ", dbErr)
			sendMessage("Ошибка создания ссылки реферала", update, tgbotapi.NewInlineKeyboardMarkup())
			return
		}
		editMessage(fmt.Sprintf("Ваша реферальная ссылка:\n\n%s", REFERRAL_LINK+ref_uuid), update, menuKeyboard)

	default:
		callback_data := update.CallbackQuery.Data
		println(callback_data)
		if strings.HasPrefix(callback_data, "user-history-topic-") {
			//Топик пользователя
			topic := strings.TrimPrefix(callback_data, "user-history-topic-")
			println(topic)
			//Получаем новости пользователя
			news, rediErr := redidb.GetNewsByTopic(topic)
			if rediErr != nil {
				log.Println("Ошибка получения новостей по топику: ", rediErr)
				sendMessage("Ошибка получения новостей", update, tgbotapi.NewInlineKeyboardMarkup())
				return
			}
			//Есть ли новости
			if len(news.News) == 0 {
				sendMessage("Новостей по данному топику нет", update, tgbotapi.NewInlineKeyboardMarkup())
				return
			}

			//Добавляем информацию о новостях пользователю
			UserHistoryMu.Lock()
			UserHistory[update.CallbackQuery.From.ID] = news
			UserHistoryCurrent[update.CallbackQuery.From.ID] = 0
			UserHistoryMu.Unlock()
			text := createNewsToText(news, 0)
			editMessage(text, update, CallbackNewsFirst)
			return
		} else if strings.HasPrefix(callback_data, "user-get-history-") {
			callback_type := strings.TrimPrefix(callback_data, "user-get-history-")
			var index int
			// Записываем текущий индекс
			UserHistoryMu.RLock()
			index = UserHistoryCurrent[update.CallbackQuery.From.ID]
			UserHistoryMu.RUnlock()
			println(index)
			println(callback_type)
			switch callback_type {
			case "next":
				index += 1
			case "prev":
				index -= 1
			}
			println(index)
			// получаем новости пользователя
			UserHistoryMu.RLock()
			user_news := UserHistory[update.CallbackQuery.From.ID]
			UserHistoryMu.RUnlock()
			UserHistoryMu.Lock()
			UserHistoryCurrent[update.CallbackQuery.From.ID] = index
			UserHistoryMu.Unlock()
			fmt.Printf("TOTAL RESULTS: %d", user_news.TotalResults)
			if index > (user_news.TotalResults-1) || index < 0 {
				println("bad index")
				return
			}
			println("good_index")
			text_to_send := createNewsToText(user_news, index)
			switch index {
			case 0:
				editMessage(text_to_send, update, CallbackNewsFirst)
				return
			case user_news.TotalResults - 1:
				editMessage(text_to_send, update, CallbackNewsLast)
				return
			default:
				editMessage(text_to_send, update, CallbackNews)
				return
			}

		}
		sendMessage("Неизвестная команда", update, tgbotapi.NewInlineKeyboardMarkup())
	}
}

func sendMessage(text string, update tgbotapi.Update, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	if len(keyboard.InlineKeyboard) > 0 {
		msg.ReplyMarkup = &keyboard
	}
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Failed to edit message: ", err)
	}
}

// Создает из NewsResponse готовый к отправке текст
func createNewsToText(news *newstypes.NewsResponse, index int) string {
	news_by_number := news.News[index]
	text := fmt.Sprintf("%s\n\n%s\n", news_by_number.Title, news_by_number.Summary)
	return text
}

func sendCallbackMessage(text string, update tgbotapi.Update, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
	if len(keyboard.InlineKeyboard) > 0 {
		msg.ReplyMarkup = &keyboard
	}
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Failed to edit message: ", err)
	}
}
func editMessage(text string, update tgbotapi.Update, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
	if len(keyboard.InlineKeyboard) > 0 {
		msg.ReplyMarkup = &keyboard
	}
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Failed to edit message: ", err)
	}
}
