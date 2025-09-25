package main

import (
	"ai_tg_search/godb"
	"ai_tg_search/redidb"
	tg_bot "ai_tg_search/tg_bot"
	"log"
	"os"

	dotenv "github.com/joho/godotenv"
)

func main() {
	//Загрузка переменных окружения из файла .env
	err := dotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	//Инициализация postgresql
	godb.InitDB()
	//Инициализация redis
	redidb.InitRedi()
	//Инициализация telegram bot
	tg_api := os.Getenv("TG_BOT_API")
	tg_bot.InitBot(tg_api)

}
