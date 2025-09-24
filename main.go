package main

import (
	tg_bot "ai_tg_search/tg_bot"
	"log"
	"os"

	dotenv "github.com/joho/godotenv"
)

func main() {
	err := dotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	tg_api := os.Getenv("TG_BOT_API")
	tg_bot.InitBot(tg_api)
}
