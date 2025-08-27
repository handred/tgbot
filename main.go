package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tgbot/libs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ .env файл не найден — используем системные переменные")
	}
}

func main() {
	log.Println("Телеграм-бот запущен...")
	go startBot() // Telegram-бот в фоне
	// Ждём сигнала завершения
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c // ← висим здесь, пока не нажмут Ctrl+C

	// Как только нажали Ctrl+C — продолжаем:
	log.Println("Телеграм-бот завершил работу...")
	// cleanup...
}

func startBot() {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		log.Fatal("Error NewBotAPI: ", err)
	}

	log.Printf("Авторизован как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Error getting updates: ", err)
	}

	for update := range updates {
		update := update // ✅ Защита от ошибки замыкания

		if update.CallbackQuery != nil {
			log.Printf("CallbackQuery: %+v", update.CallbackQuery)
		}

		if update.Message != nil {
			log.Printf("[%s] %s", getUserDisplayName(update.Message.From), update.Message.Text)

			// Создаём apiBot по требованию
			apiBot := func() *libs.ApiBot {
				return libs.NewBotApi(bot, update.Message.Chat.ID)
			}

			if update.Message.Text == "/start" {
				go apiBot().SendButtons("Добро пожаловать!")
				continue
			}

			youtubeCode := libs.GetCode(update.Message.Text)
			if youtubeCode != "" {
				videoURL := "https://www.youtube.com/watch?v=" + youtubeCode

				apiBot().SendText(fmt.Sprintf("Загружаю видео %s", videoURL))

				// Запускаем загрузку в фоне
				go func() {
					err := libs.DownloadVideo(videoURL, func(msg string) {
						// Можно логировать или отправлять прогресс пользователю
						apiBot().SendText(fmt.Sprintf("Progress: %s", msg))
						log.Println("Progress:", msg)
					})
					if err != nil {
						log.Printf("Download failed: %v", err)
						apiBot().SendText(fmt.Sprintf("Download failed: %v", err))
					}
				}()
			} else {
				go apiBot().Animate("Привет, это анимация! 🎉", 15*time.Millisecond)
			}

		}
	}
}

func getUserDisplayName(user *tgbotapi.User) string {
	if user == nil {
		return "Unknown"
	}

	if user.UserName != "" {
		return "@" + user.UserName
	}

	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	return name
}
