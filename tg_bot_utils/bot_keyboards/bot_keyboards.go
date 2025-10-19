package botkeyboards

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// Инлайн клавиатуры пользователя
//
// startKeyboard - начальная клавиатура на menu и /start
var StartKeyboard = tgbotapi.NewInlineKeyboardMarkup(
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

// Меню, предлагающее вернуться на начлаьный этап
var MenuKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
	),
)

// Профиль пользователя
var ProfileKeyboard = tgbotapi.NewInlineKeyboardMarkup(
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

// Клавиатура при просмотре истории запросов
// 0 индекс - для первой новости
// 1 индекс - основная клавиатура
// 2 индекс - для последней новости
var HistoryOfNewsKeyboard = [3]tgbotapi.InlineKeyboardMarkup{
	tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➡️", "user-get-history-next"),
		),
	),
	tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", "user-get-history-prev"),
			tgbotapi.NewInlineKeyboardButtonData("➡️", "user-get-history-next"),
		),
	),
	tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", "user-get-history-prev"),
		),
	),
}

// Клавиатура при поиске новостей через /search
// 0 индекс - для первой новости
// 1 индекс - основная клавиатура
// 2 индекс - для последней новости
var SearchNewsKeyboard = [3]tgbotapi.InlineKeyboardMarkup{
	tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➡️", "user-get-search-next"),
		),
	),
	tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", "user-get-search-prev"),
			tgbotapi.NewInlineKeyboardButtonData("➡️", "user-get-search-next"),
		),
	),
	tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️", "user-get-search-prev"),
		),
	),
}

// Первая клавиатура, не имеет опции назад
var SearchNewsFirst = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➡️", "user-get-search-next"),
	),
)

// Последняя клавиатура, не имеет опции вперед
var SearchNewsLast = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Меню", "menu"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️", "user-get-search-prev"),
	),
)
