package libs

import (
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// TextAnimator — структура для анимации текста
type ApiBot struct {
	Bot    *tgbotapi.BotAPI
	chatID int64
}

// NewTextAnimator — конструктор
func NewBotApi(bot *tgbotapi.BotAPI, chatID int64) *ApiBot {
	return &ApiBot{
		Bot:    bot,
		chatID: chatID,
	}
}

// Text — отправляет текст в чат
func (a *ApiBot) SendText(text string) {
	initialMsg := tgbotapi.NewMessage(a.chatID, text)
	a.Bot.Send(initialMsg)
}

// Buttons — отправляет кнопки в чат
func (a *ApiBot) SendButtons(text string) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Кнопка 1"),
			tgbotapi.NewKeyboardButton("Кнопка 2"),
		),
	)
	msg := tgbotapi.NewMessage(a.chatID, text)
	msg.ReplyMarkup = keyboard
	a.Bot.Send(msg)
}

// Animate — анимирует появление текста по буквам в одном сообщении
func (a *ApiBot) Animate(text string, delay time.Duration) {
	// Отправляем начальное сообщение
	initialMsg := tgbotapi.NewMessage(a.chatID, "_")
	sentMessage, err := a.Bot.Send(initialMsg)
	if err != nil {
		log.Printf("❌ Не удалось отправить начальное сообщение: %v", err)
		return
	}

	messageID := sentMessage.MessageID
	time.Sleep(100 * time.Millisecond) // даём Telegram время

	var current, lastSent string

	for _, char := range text {
		current += string(char)

		if current == lastSent {
			time.Sleep(delay)
			continue
		}

		editMsg := tgbotapi.NewEditMessageText(a.chatID, messageID, current)
		_, err := a.Bot.Send(editMsg)

		if err != nil {
			if strings.Contains(err.Error(), "message is not modified") {
				log.Printf("ℹ️ Сообщение не изменилось: '%s'", current)
			} else {
				log.Printf("❌ Ошибка редактирования: %v", err)
			}
		} else {
			lastSent = current
		}

		time.Sleep(delay)
	}
}
