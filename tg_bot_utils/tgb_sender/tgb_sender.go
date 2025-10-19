//! tgbsender отвечает за отправку, либо же редактирование сообщений в телеграмм боте

package tgbsender

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Базовая отправка пользователю сообщения на его действия
func SendMessage(bot *tgbotapi.BotAPI, text string, update tgbotapi.Update, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	if len(keyboard.InlineKeyboard) > 0 {
		msg.ReplyMarkup = &keyboard
	}
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Failed to edit message: ", err)
	}
}

func SendMessageByUserID(bot *tgbotapi.BotAPI, text string, userID int64, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(userID, text)
	if len(keyboard.InlineKeyboard) > 0 {
		msg.ReplyMarkup = &keyboard
	}
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Failed to edit message: ", err)
	}
}

func SendCallbackMessage(bot *tgbotapi.BotAPI, text string, update tgbotapi.Update, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
	if len(keyboard.InlineKeyboard) > 0 {
		msg.ReplyMarkup = &keyboard
	}
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Failed to edit message: ", err)
	}
}
func EditCallbackMessage(bot *tgbotapi.BotAPI, text string, update tgbotapi.Update, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, text)
	if len(keyboard.InlineKeyboard) > 0 {
		msg.ReplyMarkup = &keyboard
	}
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Failed to edit message: ", err)
	}
}

func EditChatMessage(bot *tgbotapi.BotAPI, text string, update tgbotapi.Update, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewEditMessageText(update.Message.MigrateFromChatID, update.Message.MessageID, text)
	if len(keyboard.InlineKeyboard) > 0 {
		msg.ReplyMarkup = &keyboard
	}
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Failed to edit message: ", err)
	}
}
