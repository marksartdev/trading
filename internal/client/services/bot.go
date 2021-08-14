package services

import (
	"fmt"
	"strconv"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/marksartdev/trading/internal/client"
	"github.com/marksartdev/trading/internal/client/delivery/rpc"
	"github.com/marksartdev/trading/internal/config"
	"github.com/marksartdev/trading/internal/log"
)

const botAction log.Action = "bot"

// Telegram bot.
type telegramBot struct {
	logger log.Logger
	cfg    config.Client
	broker rpc.BrokerService
	bot    *tgbotapi.BotAPI
	chats  map[int64]actionPlane
}

// NewTelegramBot creates new telegram bot.
func NewTelegramBot(logger log.Logger, cfg config.Client, broker rpc.BrokerService) client.TelegramBot {
	return &telegramBot{logger: logger, cfg: cfg, broker: broker, chats: make(map[int64]actionPlane)}
}

// Start starts bot.
func (t *telegramBot) Start() error {
	bot, err := tgbotapi.NewBotAPI(t.cfg.Token)
	if err != nil {
		return err
	}

	t.bot = bot

	t.logger.Info(botAction, "started")

	updCfg := tgbotapi.NewUpdate(0)
	updCfg.Timeout = 60

	updates, err := t.bot.GetUpdatesChan(updCfg)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		t.logger.Info(botAction, fmt.Sprintf("[%s] %s", update.Message.From.UserName, update.Message.Text))

		switch update.Message.Text {
		case "/create":
			t.chats[update.Message.Chat.ID] = createPlan()
			t.input(update.Message.Chat.ID, "")
		case "/cancel":
			t.chats[update.Message.Chat.ID] = cancelPlan()
			t.input(update.Message.Chat.ID, "")
		case "/profile":
			t.profile(update.Message.Chat.ID, int64(update.Message.From.ID))
		case "/statistic":
			t.chats[update.Message.Chat.ID] = statPlan()
			t.input(update.Message.Chat.ID, "")
		default:
			if t.input(update.Message.Chat.ID, update.Message.Text) {
				switch t.chats[update.Message.Chat.ID].action {
				case create:
					t.create(update.Message.Chat.ID, int64(update.Message.From.ID))
				case cancel:
					t.cancel(update.Message.Chat.ID, int64(update.Message.From.ID))
				case statistic:
					t.statistic(update.Message.Chat.ID, int64(update.Message.From.ID))
				}
			}
		}
	}

	return nil
}

func (t *telegramBot) create(chatID, userID int64) {
	defer delete(t.chats, chatID)
	chat := t.chats[chatID]

	login := t.getLogin(userID)

	dealType := chat.answers[1]
	if dealType != "BUY" && dealType != "SELL" {
		t.sendMsg(chatID, "Вы ввели некорректный тип сделки")
		t.sendMsg(chatID, "Возможны только SELL и BUY")
		t.sendMsg(chatID, "Придется начать сначала =(")
		return
	}

	amn, err := strconv.Atoi(chat.answers[2])
	if err != nil {
		t.handleErr(chatID, err)
		return
	}

	price, err := strconv.ParseFloat(chat.answers[3], 64)
	if err != nil {
		t.handleErr(chatID, err)
		return
	}

	msg, err := t.broker.Create(login, chat.answers[0], chat.answers[1], int32(amn), price)
	if err != nil {
		t.handleErr(chatID, err)
		return
	}

	t.sendMsg(chatID, msg)
}

func (t *telegramBot) cancel(chatID, userID int64) {
	defer delete(t.chats, chatID)
	chat := t.chats[chatID]

	login := t.getLogin(userID)

	dealID, err := strconv.ParseInt(chat.answers[0], 10, 64)
	if err != nil {
		t.handleErr(chatID, err)
		return
	}

	msg, err := t.broker.Cancel(login, dealID)
	if err != nil {
		t.handleErr(chatID, err)
		return
	}

	t.sendMsg(chatID, msg)
}

func (t *telegramBot) profile(chatID, userID int64) {
	login := t.getLogin(userID)
	msg, err := t.broker.Profile(login)
	if err != nil {
		t.handleErr(chatID, err)
		return
	}

	t.sendMsg(chatID, msg)
}

func (t *telegramBot) statistic(chatID, userID int64) {
	defer delete(t.chats, chatID)
	chat := t.chats[chatID]

	login := t.getLogin(userID)
	msg, err := t.broker.Statistic(login, chat.answers[0])
	if err != nil {
		t.handleErr(chatID, err)
		return
	}

	t.sendMsg(chatID, msg)
}

func (t *telegramBot) input(chatID int64, msg string) bool {
	chat, ok := t.chats[chatID]
	if !ok {
		return false
	}
	defer func() {
		t.chats[chatID] = chat
	}()

	if msg != "" {
		chat.answers[chat.ptr] = msg
		chat.ptr++
	}

	if chat.ptr == len(chat.questions) {
		return true
	}

	t.sendMsg(chatID, chat.questions[chat.ptr])
	return false
}

func (t *telegramBot) handleErr(chatID int64, err error) {
	t.logger.Error(botAction, err)
	t.sendMsg(chatID, "Упс! Что-то пошло не так =(")
}

func (t *telegramBot) sendMsg(chatID int64, msg string) {
	message := tgbotapi.NewMessage(chatID, msg)
	if _, err := t.bot.Send(message); err != nil {
		t.logger.Error(botAction, err)
	}
}

func (t *telegramBot) getLogin(userID int64) string {
	return fmt.Sprintf("tg-%d", userID)
}
