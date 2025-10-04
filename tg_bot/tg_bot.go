package tgbot

import (
	godb "ai_tg_search/godb"
	"fmt"
	"log"
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

// Инлайн клавиатуры пользователя
//
// startKeyboard
var startKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Начать поиск", "search"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Помощь", "help"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Реферал", "refferal"),
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
	case "search":
		UserStateMu.Lock()
		UserState[update.CallbackQuery.From.ID] = "user-search"
		UserStateMu.Unlock()
		editMessage("Введите запрос для поиска новостей\n\nЗапросом должен являться текст, по которому будет выполнен поиск новостей\nПример: Рынок криптовалют, лучшие тенденции, восходящие альт коины", update, tgbotapi.NewInlineKeyboardMarkup())
	case "help":
		sendMessage("Добро пожаловать в наш бот!\nОн создан специально для быстрого поиска новостей.\n\nНажмите ниже на inline кнопку", update, tgbotapi.NewInlineKeyboardMarkup())
	case "refferal":
		ref_uuid, dbErr := godb.GetUserReferralLink(update.CallbackQuery.From.ID)
		if dbErr != nil {
			log.Println("Ошибка поиска реферальной ссылки: ", dbErr)
			sendMessage("Ошибка создания ссылки реферала", update, tgbotapi.NewInlineKeyboardMarkup())
			return
		}
		editMessage(fmt.Sprintf("Ваша реферальная ссылка:\n\n%s", REFERRAL_LINK+ref_uuid), update, tgbotapi.NewInlineKeyboardMarkup())

	default:
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
