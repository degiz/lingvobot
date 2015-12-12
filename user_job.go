package main

import (
	"encoding/json"
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
	ivona "github.com/jpadilla/ivona-go"
	"gopkg.in/redis.v3"
	"log"
	"math/rand"
	"time"
)

const (
	helpMessage = `Hey there, I'm Lingvo Bot
Here are my commands:
/start is posting you words
/stop stops posting
/go asks you questions
/audio send you pronunciation
/help prints this message
Enjoy!`

	stopMessage = `Ok, no more words.
Press /start to begin again!`

	audioMessage = "Ok, your next message will be traslated into the audio file!"

	readyMessage     = "Press /go once you're ready"
	correctMessage   = "Genau!"
	incorrectMessage = "Nein :("
)

const (
	waitingForStart  = iota
	waitingForGo     = iota
	waitingForAnswer = iota
	waitingForAudio  = iota
)

type UserJob struct {
	state            *UserState
	controls         *Contols
	chatID           int
	currentNouns     []Noun
	currentWordID    int
	numOfQuestions   int
	keyboard         tgbotapi.ReplyKeyboardMarkup
	waitingState     int
	prevWaitingState int
}

func (job *UserJob) Start() {

	job.init()

	if !job.CheckUserExists() {
		job.SendMessage(helpMessage)
		job.SaveState()
	}

	for {
		select {
		case message := <-job.state.channel:
			job.chatID = message.Chat.ID
			job.ProcessMessage(message.Text)
		case <-job.state.quit:
			log.Printf("Goroutine is going down..")
			return
		}
	}
}

func (job *UserJob) init() {
	job.numOfQuestions = 3
	job.keyboard = tgbotapi.ReplyKeyboardMarkup{}

	buttons := [][]string{[]string{"der", "die"}, []string{"das", "die (pl)"}}

	job.keyboard.OneTimeKeyboard = true
	job.keyboard.ResizeKeyboard = true
	job.keyboard.Keyboard = buttons

	job.waitingState = waitingForStart
}

func (job *UserJob) ProcessMessage(message string) {

	switch message {
	case "/help":
		job.SendMessage(helpMessage)
		return
	case "/stop":
		job.SendMessage(stopMessage)
		return
	case "/audio":
		job.SendMessage(audioMessage)
		job.SaveWaitingState()
		job.waitingState = waitingForAudio
		return
	}

	switch job.waitingState {
	case waitingForStart:
		if message == "/start" {
			job.SendNewWords()
		} else {
			job.SendMessage(helpMessage)
		}
	case waitingForGo:
		if message == "/go" {
			job.SendSticker()
			job.SendQuestion()
		} else {
			job.SendMessage(readyMessage)
		}
	case waitingForAnswer:
		job.CheckAnswers(message)
	case waitingForAudio:
		job.SendAudio(message)
	}

}

func (job *UserJob) SendNewWords() {
	job.currentNouns = make([]Noun, 0)
	job.currentWordID = 0
	ids := getSampleNumbers(job.controls.nounKeys, job.numOfQuestions)
	var noun Noun
	for _, id := range ids {
		nounJSON, _ := job.controls.redis.Get(getNounIDKey(id)).Result()
		err := json.Unmarshal([]byte(nounJSON), &noun)
		if err != nil {
			log.Printf("Cannot unmarshal: %s", err)
		}
		message := fmt.Sprintf("%s %s - %s", noun.Article, noun.Noun, noun.Translation)
		job.SendMessage(message)
		job.currentNouns = append(job.currentNouns, noun)
	}
	job.waitingState = waitingForGo
	job.SendMessage(readyMessage)
}

func (job *UserJob) SendQuestion() {
	text := fmt.Sprintf("%s (%s)",
		job.currentNouns[job.currentWordID].Noun,
		job.currentNouns[job.currentWordID].Translation)
	message := tgbotapi.NewMessage(job.chatID, text)

	message.ReplyMarkup = job.keyboard

	job.waitingState = waitingForAnswer
	job.controls.telegramBot.Send(message)
}

func (job *UserJob) CheckAnswers(message string) {
	if message == job.currentNouns[job.currentWordID].Article {
		job.SendMessage(correctMessage)
		if job.currentWordID < job.numOfQuestions-1 {
			job.currentWordID++
			job.SendQuestion()
		} else {
			job.SendNewWords()
		}
	} else {
		job.SendMessage(incorrectMessage)
		job.SendQuestion()
	}
}

func (job *UserJob) SendMessage(text string) {
	message := tgbotapi.NewMessage(job.chatID, text)
	job.controls.telegramBot.Send(message)
}

func (job *UserJob) SendSticker() {
	// message := tgbotapi.NewStickerUpload(chatID, file)
	// job.controls.telegramBot.Send(message)
}

func (job *UserJob) SendAudio(message string) {
	options := ivona.NewSpeechOptions(message)
	options.Voice.Name = "Marlene"
	options.Voice.Language = "de-DE"
	options.Parameters.Rate = "slow"
	response, err := job.controls.ivona.CreateSpeech(options)

	if err != nil {
		log.Printf("Cannot get audio file: %s", err)
		return
	}

	msg := tgbotapi.NewAudioUpload(
		job.chatID,
		tgbotapi.FileBytes{Name: "audio.mp3", Bytes: response.Audio},
	)

	job.controls.telegramBot.Send(msg)
	job.RestoreWaitingState()
}

func (job *UserJob) SaveState() {
	job.controls.redis.Set(job.state.getIDString(), "0", 0)
}

func (job *UserJob) CheckUserExists() bool {
	_, err := job.controls.redis.Get(job.state.getIDString()).Result()
	if err == redis.Nil {
		return false
	}
	return true
}

func (job *UserJob) SaveWaitingState() {
	job.prevWaitingState = job.waitingState
}

func (job *UserJob) RestoreWaitingState() {
	job.waitingState = job.prevWaitingState
}

// Reservoir sampling
func getSampleNumbers(keys []int, reservoirSize int) []int {
	result := make([]int, reservoirSize)

	rand.Seed(time.Now().UTC().UnixNano())

	for i := 0; i < reservoirSize; i++ {
		result[i] = keys[i]
	}

	for i := reservoirSize; i < len(keys); i++ {
		r := rand.Intn(i + 1)
		if r < reservoirSize {
			result[r] = keys[i]
		}
	}

	return result
}
