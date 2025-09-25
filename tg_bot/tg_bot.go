package tgbot

import (
	"log"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

func InitBot(token string) {
	bot_instant, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}
	bot = bot_instant
	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)

	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)

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
		args := update.Message.CommandArguments()
		if args == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать в наш бот!\nОн создан специально для быстрого поиска новостей.\n\nНажмите ниже на inline кнопку")
			msg.ReplyMarkup = startKeyboard
			sendMessage(msg)
		} else if strings.HasPrefix(args, "ref_") {
			referalUUID := strings.TrimPrefix(args, "ref_")
			log.Println(referalUUID)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Добро пожаловать в наш бот!\nОн создан специально для быстрого поиска новостей.\n\nНажмите ниже на inline кнопку")
			msg.ReplyMarkup = startKeyboard
			sendMessage(msg)
		}

	case "help":

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
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Запрос успешно сохранен. Ожидайте ответа от бота.")
		sendMessage(msg)

	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда")
		sendMessage(msg)
	}
}

func handleCallback(update tgbotapi.Update) {
	switch update.CallbackQuery.Data {
	case "search":
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Введите запрос для поиска новостей\n\nЗапросом должен являться текст, по которому будет выполнен поиск новостей\nПример: Рынок криптовалют, лучшие тенденции, восходящие альт коины")
		UserStateMu.Lock()
		UserState[update.CallbackQuery.From.ID] = "user-search"
		UserStateMu.Unlock()
		editMessage(msg)
	case "help":
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Добро пожаловать в наш бот!\nОн создан специально для быстрого поиска новостей.\n\nНажмите ниже на inline кнопку")
		msg.ReplyMarkup = startKeyboard
		sendMessage(msg)
	default:
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Неизвестная команда")
		sendMessage(msg)
	}
}

func sendMessage(msg tgbotapi.MessageConfig) {
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Failed to edit message: ", err)
	}
}

func editMessage(msg tgbotapi.EditMessageTextConfig) {
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Failed to edit message: ", err)
	}
}
