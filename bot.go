package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	ivona "github.com/jpadilla/ivona-go"
	"gopkg.in/redis.v3"
	"io"
	"log"
	"os"
	"time"
)

type Contols struct {
	redis       *redis.Client
	telegramBot *tgbotapi.BotAPI
	ivona       *ivona.Ivona
	nounKeys    []int
}

type Bot struct {
	configFile        string
	config            *Config
	controls          *Contols
	activeUsersStates map[int]*UserState
}

// TODO: read messages from config
const (
	stopCommand = "/stop"
)

func (bot *Bot) Run() error {

	var configErr error
	bot.config, configErr = getConfig(bot.configFile)
	if configErr != nil {
		panic(fmt.Sprintf("Failed to read config: %v", configErr))
	}

	bot.controls = &Contols{}

	err := bot.setUpRedis()
	if err != nil {
		panic(fmt.Sprintf("Redis init failed: %v", err))
	}

	err = bot.loadWords()
	if err != nil {
		panic(fmt.Sprintf("Loading words failed: %v", err))
	}

	err = bot.setUpTelegramBot()
	if err != nil {
		panic(fmt.Sprintf("Telegram bot init failed: %v", err))
	}

	err = bot.setUpTTS()
	if err != nil {
		panic(fmt.Sprintf("TTS init failed: %v", err))
	}

	bot.activeUsersStates = make(map[int]*UserState)

	// TODO: save the offset
	// TODO: read timeout from config
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 1

	updates, _ := bot.controls.telegramBot.GetUpdatesChan(u)

	for update := range updates {
		bot.processUpdate(update)
	}

	return nil
}

func (bot *Bot) setUpRedis() error {

	var err error
	attempts := 10

	for attempts > 0 {
		bot.controls.redis = redis.NewClient(&redis.Options{
			Addr:     bot.config.RedisAddress,
			Password: bot.config.RedisPassword,
			DB:       bot.config.RedisDb,
		})

		_, err = bot.controls.redis.Ping().Result()
		if err != nil {
			log.Printf("Retrying to connect to redis..")
			time.Sleep(1 * time.Second)
		} else {
			break
		}
		attempts--
	}

	return err
}

func (bot *Bot) loadWords() error {
	nouns, err := readNounsFromFile(bot.config.NounsPath)
	if err != nil {
		return err
	}
	bot.controls.nounKeys = make([]int, 0, len(nouns))

	// Populate redis with nouns
	for _, noun := range nouns {
		bot.controls.nounKeys = append(bot.controls.nounKeys, noun.ID)
		jsonNoun, err := json.Marshal(noun)
		if err != nil {
			return err
		}
		bot.controls.redis.Set(noun.getIDString(), jsonNoun, 0)
	}

	return nil
}

func readNounsFromFile(filename string) (map[int]*Noun, error) {

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvr := csv.NewReader(f)

	id := 0
	nouns := make(map[int]*Noun)

	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return nouns, err
		}

		noun := &Noun{}

		noun.ID = id
		noun.Translation = row[0]
		noun.Article = row[1]
		noun.Noun = row[2]

		nouns[id] = noun
		id++
	}

	return nouns, nil
}

func (bot *Bot) setUpTelegramBot() error {
	var err error
	bot.controls.telegramBot, err = tgbotapi.NewBotAPI(bot.config.TelegramBotToken)
	if err != nil {
		return err
	}

	return nil
}

func (bot *Bot) setUpTTS() error {
	bot.controls.ivona = ivona.New(
		bot.config.IvonaAccessKeyToken,
		bot.config.IvonaSecretKeyToken)
	return nil
}

func (bot *Bot) processUpdate(update tgbotapi.Update) error {

	userID := update.Message.From.ID

	if _, ok := bot.activeUsersStates[userID]; !ok {
		err := bot.ProcessNewUser(userID, update.Message.Chat.ID)
		if err != nil {
			panic(fmt.Sprintf("Failed processing new user: %v", err))
		}
	}

	bot.activeUsersStates[userID].channel <- update.Message

	if update.Message.Text == stopCommand {
		err := bot.RemoveUser(userID)
		if err != nil {
			panic(fmt.Sprintf("Failed removing user: %v", err))
		}
	}

	return nil
}

func (bot *Bot) ProcessNewUser(userID int, chatID int) error {

	log.Printf("Processing new user with id: %d", userID)
	userState := &UserState{}
	userState.channel = make(chan tgbotapi.Message)
	userState.quit = make(chan int)
	userState.UserID = userID

	bot.activeUsersStates[userID] = userState

	job := &UserJob{state: userState, controls: bot.controls, chatID: chatID}
	go job.Start()

	return nil
}

func (bot *Bot) RemoveUser(userID int) error {

	log.Printf("Removing user with id: %d", userID)
	bot.activeUsersStates[userID].quit <- 1
	delete(bot.activeUsersStates, userID)

	return nil
}
