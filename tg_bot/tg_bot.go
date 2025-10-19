package tgbot

import (
	godb "ai_tg_search/godb"
	botkeyboards "ai_tg_search/tg_bot_utils/bot_keyboards"
	historycallback "ai_tg_search/tg_bot_utils/callback_handler/history_callback"
	newscallback "ai_tg_search/tg_bot_utils/callback_handler/news_callback"
	profilecallback "ai_tg_search/tg_bot_utils/callback_handler/profile_callback"
	tgbsender "ai_tg_search/tg_bot_utils/tgb_sender"
	userstates "ai_tg_search/tg_bot_utils/user_states"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const REFERRAL_LINK = "https://t.me/newsly_search_bot?start=ref_"

var bot *tgbotapi.BotAPI

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
	log.Printf("\n------------\nBot started his job with username %s\n------------", bot.Self.UserName)
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
		//Проверка наличия реферальной ссылки
		var refferal_string string
		args := update.Message.CommandArguments()
		if len(args) > 0 && strings.TrimPrefix(args, "ref_") != "" {
			refferal_string = strings.Split(args, "_")[1]
		}
		dbErr := godb.AddUserIfNotExists(update.Message.From.ID, refferal_string)
		if dbErr != nil {
			log.Println("Error adding user:", dbErr)
			tgbsender.SendMessage(bot, "Ошибка при добавлении пользователя. Пожалуйста, попробуйте позже.", update, tgbotapi.NewInlineKeyboardMarkup())
		}
		tgbsender.SendMessage(bot, "Добро пожаловать в newsly!\n\nБыстрый и удобный бот для поиска новостей на любую тематику.\n\nВыберите опцию ниже", update, botkeyboards.StartKeyboard)

	case "help":
		tgbsender.SendMessage(bot, "Для поиска новостей введите команду /search и тему, например: /search IT", update, botkeyboards.MenuKeyboard)
	case "search":
		newscallback.HandleSearchCommand(bot, &update)
	case "add":
		text := strings.Split(update.Message.CommandArguments(), " ")
		if len(text) > 1 {
			//Парсинг аргументов
			user_id := text[0]
			amount := text[1]
			//перевод user_id в int
			user_int_id, IDerr := strconv.Atoi(user_id)
			if IDerr != nil {
				tgbsender.SendMessage(bot, "Введите полноценный ID пользователя", update, tgbotapi.NewInlineKeyboardMarkup())
			}
			// int -> int64
			user_int64_id := int64(user_int_id)
			// перевод amount в int64
			amount_int, amountErr := strconv.Atoi(amount)
			if amountErr != nil {
				tgbsender.SendMessage(bot, "Введите полноценную сумму", update, tgbotapi.NewInlineKeyboardMarkup())
			}
			dbErr := godb.AddRequests(update.Message.From.ID, user_int64_id, amount_int)
			if dbErr != nil {
				log.Println(dbErr)
				tgbsender.SendMessage(bot, "Ошибка при добавлении запроса", update, tgbotapi.NewInlineKeyboardMarkup())
			}
			tgbsender.SendMessage(bot, "Запрос успешно добавлен", update, tgbotapi.NewInlineKeyboardMarkup())
			tgbsender.SendMessageByUserID(bot, fmt.Sprintf("Вам было успешно начислено %d запрсов!\n\nПряитного пользования!", amount_int), user_int64_id, botkeyboards.MenuKeyboard)
		}
	default:
		tgbsender.SendMessage(bot, "Такой команды у нас нет!", update, botkeyboards.MenuKeyboard)
	}
}

func handleMessage(update tgbotapi.Update) {
	state := userstates.UserState[update.Message.Chat.ID]
	switch state {
	case "user-search":
		println("make it already, ok?")
	default:
		tgbsender.SendMessage(bot, "Неизвестная команда", update, tgbotapi.NewInlineKeyboardMarkup())
	}
}

func handleCallback(update tgbotapi.Update) {
	defer func() {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Обработано")
		if _, err := bot.Request(callback); err != nil {
			log.Println("Ошибка при отправке callback:", err)
		}
	}()
	switch update.CallbackQuery.Data {
	case "menu":
		tgbsender.EditCallbackMessage(bot, "Добро пожаловать в newsly!\n\nБыстрый и удобный бот для поиска новостей на любую тематику.\n\nВыберите опцию ниже", update, botkeyboards.StartKeyboard)
	case "profile":
		profilecallback.HandleProfileCallback(bot, update)
	case "history":
		historycallback.HandleHistoryCallback(bot, update)
	case "help":
		tgbsender.SendCallbackMessage(bot, "Добро пожаловать в наш бот!\nОн создан специально для быстрого поиска новостей.\n\nЧтобы начать поиск новостей введите команду\n\n/search <Требуемая новость>", update, botkeyboards.MenuKeyboard)
	case "refferal":
		ref_uuid, dbErr := godb.GetUserReferralLink(update.CallbackQuery.From.ID)
		if dbErr != nil {
			log.Println("Ошибка поиска реферальной ссылки: ", dbErr)
			tgbsender.SendCallbackMessage(bot, "Ошибка создания ссылки реферала", update, tgbotapi.NewInlineKeyboardMarkup())
			return
		}
		tgbsender.EditCallbackMessage(bot, fmt.Sprintf("Ваша реферальная ссылка:\n\n%s", REFERRAL_LINK+ref_uuid), update, botkeyboards.MenuKeyboard)

	default:
		callback_data := update.CallbackQuery.Data
		println(callback_data)
		if strings.HasPrefix(callback_data, "user-history-topic-") {
			historycallback.HandleHistoryTopicCallback(bot, update)
			return
		} else if strings.HasPrefix(callback_data, "user-get-") {
			newscallback.HandleSearchOrHistoryCallback(bot, &update)
		} else {
			tgbsender.SendMessage(bot, "Неизвестная команда", update, tgbotapi.NewInlineKeyboardMarkup())
		}
	}
}
