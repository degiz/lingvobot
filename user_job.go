package main

import (
	"encoding/json"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	"gopkg.in/redis.v3"
	"log"
	"math/rand"
	"time"
)

const (
	help_message = `Hey there, I'm Lingvo Bot
Here are my commands:
/start is posting you words
/stop stops posting
/go asks you questions
/help prints this message
Enjoy!`

	stop_message = `Ok, no more words.
Press /start to begin again!`

	ready_message     = "Press /go once you're ready"
	correct_message   = "Genau!"
	incorrect_message = "Nein :("
)

const (
	waiting_for_start  = iota
	waiting_for_go     = iota
	waiting_for_answer = iota
)

type UserJob struct {
	state            *UserState
	controls         *Contols
	chat_id          int
	current_nouns    []Noun
	current_word_id  int
	num_of_questions int
	keyboard         tgbotapi.ReplyKeyboardMarkup
	waiting_state    int
}

func (self *UserJob) Start() {

	self.init()

	if !self.CheckUserExists() {
		self.SendMessage(help_message)
		self.SaveState()
	}

	for {
		select {
		case message := <-self.state.channel:
			self.chat_id = message.Chat.ID
			self.ProcessMessage(message.Text)
		case <-self.state.quit:
			log.Printf("Goroutine is going down..")
			return
		}
	}
}

func (self *UserJob) init() {
	self.num_of_questions = 3
	self.keyboard = tgbotapi.ReplyKeyboardMarkup{}

	buttons := [][]string{[]string{"der", "die"}, []string{"das", "die (pl)"}}

	self.keyboard.OneTimeKeyboard = true
	self.keyboard.ResizeKeyboard = true
	self.keyboard.Keyboard = buttons

	self.waiting_state = waiting_for_start
}

func (self *UserJob) ProcessMessage(message string) {

	if message == "/help" {
		self.SendMessage(help_message)
		return
	} else if message == "/stop" {
		self.SendMessage(stop_message)
		return
	}

	switch self.waiting_state {
	case waiting_for_start:
		if message == "/start" {
			self.SendNewWords()
		} else {
			self.SendMessage(help_message)
		}
	case waiting_for_go:
		if message == "/go" {
			self.SendSticker()
			self.SendQuestion()
		} else {
			self.SendMessage(ready_message)
		}
	case waiting_for_answer:
		self.CheckAnswers(message)
	}

}

func (self *UserJob) SendNewWords() {
	self.current_nouns = make([]Noun, 0)
	self.current_word_id = 0
	ids := getSampleNumbers(self.controls.noun_keys, self.num_of_questions)
	var noun Noun
	for _, id := range ids {
		noun_json, _ := self.controls.redis.Get(getNounIdKey(id)).Result()
		err := json.Unmarshal([]byte(noun_json), &noun)
		if err != nil {
			log.Printf("Cannot unmarshal: %s", err)
		}
		message := fmt.Sprintf("%s %s - %s", noun.Article, noun.Noun, noun.Translation)
		self.SendMessage(message)
		self.current_nouns = append(self.current_nouns, noun)
	}
	self.waiting_state = waiting_for_go
	self.SendMessage(ready_message)
}

func (self *UserJob) SendQuestion() {
	text := fmt.Sprintf("%s (%s)",
		self.current_nouns[self.current_word_id].Noun,
		self.current_nouns[self.current_word_id].Translation)
	message := tgbotapi.NewMessage(self.chat_id, text)

	message.ReplyMarkup = self.keyboard

	self.waiting_state = waiting_for_answer
	self.controls.telegram_bot.Send(message)
}

func (self *UserJob) CheckAnswers(message string) {
	if message == self.current_nouns[self.current_word_id].Article {
		self.SendMessage(correct_message)
		if self.current_word_id < self.num_of_questions-1 {
			self.current_word_id += 1
			self.SendQuestion()
		} else {
			self.SendNewWords()
		}
	} else {
		self.SendMessage(incorrect_message)
		self.SendQuestion()
	}
}

func (self *UserJob) SendMessage(text string) {
	message := tgbotapi.NewMessage(self.chat_id, text)
	self.controls.telegram_bot.Send(message)
}

func (self *UserJob) SendSticker() {
	// message := tgbotapi.NewStickerUpload(chatID, file)
	// self.controls.telegram_bot.Send(message)
}

func (self *UserJob) SaveState() {
	self.controls.redis.Set(self.state.getIdString(), "0", 0)
}

func (self *UserJob) CheckUserExists() bool {
	_, err := self.controls.redis.Get(self.state.getIdString()).Result()
	if err == redis.Nil {
		return false
	} else {
		return true
	}
}

// Reservoir sampling
func getSampleNumbers(keys []int, reservoir_size int) []int {
	result := make([]int, reservoir_size)

	rand.Seed(time.Now().UTC().UnixNano())

	for i := 0; i < reservoir_size; i++ {
		result[i] = keys[i]
	}

	for i := reservoir_size; i < len(keys); i++ {
		r := rand.Intn(i + 1)
		if r < reservoir_size {
			result[r] = keys[i]
		}
	}

	return result
}
