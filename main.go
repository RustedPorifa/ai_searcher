package main

import (
	"ai_tg_search/godb"
	perplexityapi "ai_tg_search/perplexity_api"
	"ai_tg_search/redidb"
	tg_bot "ai_tg_search/tg_bot"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	dotenv "github.com/joho/godotenv"
)

func main() {
	//Загрузка переменных окружения из файла .env
	err := dotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	adminIDs, adminErr := getAdminIDs()
	if adminErr != nil {
		log.Fatal(adminErr)
	}

	//Инициализация postgresql
	godb.InitDB()
	//Инициализация redis
	redidb.InitRedi()
	//Инициализация perplexity api
	perplexityapi.InitPerplexityAPI()
	//Инициализация базы данных администраторов
	dbErr := godb.AddAdminsIfNotExist(adminIDs)
	if dbErr != nil {
		log.Fatal(dbErr)
	}

	//Инициализация telegram bot
	tg_api := os.Getenv("TG_BOT_API")
	tg_bot.InitBot(tg_api)

}

func getAdminIDs() ([]int64, error) {
	adminsEnv := os.Getenv("ADMIN_IDS")
	if adminsEnv == "" {
		return nil, fmt.Errorf("переменная окружения ADMIN_IDS не установлена")
	}

	adminIDsStr := strings.Split(adminsEnv, ",")
	var adminIDs []int64

	for _, idStr := range adminIDsStr {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue // игнорируем пустые элементы
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ошибка конвертации %q в int64: %w", idStr, err)
		}
		adminIDs = append(adminIDs, id)
	}

	return adminIDs, nil
}
