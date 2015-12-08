package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"gopkg.in/redis.v3"
	"io"
	"log"
	"os"
	"time"
)

type Contols struct {
	redis        *redis.Client
	telegram_bot *tgbotapi.BotAPI
	noun_keys    []int
}

type Bot struct {
	config_file         string
	config              *Config
	controls            *Contols
	active_users_states map[int]*UserState
}

// TODO: read messages from config
const (
	stop_command = "/stop"
)

func (self *Bot) Run() error {

	var config_err error
	self.config, config_err = getConfig(self.config_file)
	if config_err != nil {
		panic(fmt.Sprintf("Failed to read config: %v", config_err))
	}

	self.controls = &Contols{}

	err := self.setUpRedis()
	if err != nil {
		panic(fmt.Sprintf("Redis init failed: %v", err))
	}

	err = self.loadWords()
	if err != nil {
		panic(fmt.Sprintf("Loading words failed: %v", err))
	}

	err = self.setUpTelegramBot()
	if err != nil {
		panic(fmt.Sprintf("Telegram bot init failed: %v", err))
	}

	self.active_users_states = make(map[int]*UserState)

	// TODO: save the offset
	// TODO: read timeout from config
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 1

	updates, _ := self.controls.telegram_bot.GetUpdatesChan(u)

	for update := range updates {
		self.processUpdate(update)
	}

	return nil
}

func (self *Bot) setUpRedis() error {

	var err error
	attempts := 10

	for attempts > 0 {
		self.controls.redis = redis.NewClient(&redis.Options{
			Addr:     self.config.RedisAddress,
			Password: self.config.RedisPassword,
			DB:       self.config.RedisDb,
		})

		_, err = self.controls.redis.Ping().Result()
		if err != nil {
			log.Printf("Retrying to connect to redis..")
			time.Sleep(1 * time.Second)
		} else {
			break
		}
		attempts -= 1
	}

	return err
}

func (self *Bot) loadWords() error {
	// TODO: read path from config
	nouns, err := readNounsFromFile(self.config.NounsPath)
	if err != nil {
		return err
	}
	self.controls.noun_keys = make([]int, 0, len(nouns))

	// Populate redis with nouns
	for _, noun := range nouns {
		self.controls.noun_keys = append(self.controls.noun_keys, noun.Id)
		json_noun, err := json.Marshal(noun)
		if err != nil {
			return err
		}
		self.controls.redis.Set(noun.getIdString(), json_noun, 0)
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

		noun.Id = id
		noun.Translation = row[0]
		noun.Article = row[1]
		noun.Noun = row[2]

		nouns[id] = noun
		id += 1
	}

	return nouns, nil
}

func (self *Bot) setUpTelegramBot() error {
	var err error
	self.controls.telegram_bot, err = tgbotapi.NewBotAPI(self.config.TelegramBotToken)
	if err != nil {
		return err
	}

	return nil
}

func (self *Bot) processUpdate(update tgbotapi.Update) error {

	user_id := update.Message.From.ID

	if _, ok := self.active_users_states[user_id]; !ok {
		err := self.ProcessNewUser(user_id, update.Message.Chat.ID)
		if err != nil {
			panic(fmt.Sprintf("Failed processing new user: %v", err))
		}
	}

	self.active_users_states[user_id].channel <- update.Message

	if update.Message.Text == stop_command {
		err := self.RemoveUser(user_id)
		if err != nil {
			panic(fmt.Sprintf("Failed removing user: %v", err))
		}
	}

	return nil
}

func (self *Bot) ProcessNewUser(user_id int, chat_id int) error {

	log.Printf("Processing new user with id: %d", user_id)
	user_state := &UserState{}
	user_state.channel = make(chan tgbotapi.Message)
	user_state.quit = make(chan int)
	user_state.UserId = user_id

	self.active_users_states[user_id] = user_state

	job := &UserJob{state: user_state, controls: self.controls, chat_id: chat_id}
	go job.Start()

	return nil
}

func (self *Bot) RemoveUser(user_id int) error {

	log.Printf("Removing user with id: %d", user_id)
	self.active_users_states[user_id].quit <- 1
	delete(self.active_users_states, user_id)

	return nil
}
