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
/train starts nouns training
/audio sends you pronunciation
/noun sends you noun info
/verb sends you verb info
/help prints this message
Enjoy!`

	trainingMessage = `Here are training commands:
/start starts training
/check checks your knowledges
/stop stops training`

	stopMessage = "Ok, no more words!"

	audioMessage = "Ok, your next message will be traslated into the audio file!"

	nounMessage = "Ok, send me noun you'd like to learn!"

	noNounMessage = "Sorry, didn't find anything.."

	verbMessage = "Sorry, verbs are not yet implemented.."

	readyMessage     = "Press /check once you're ready"
	correctMessage   = "Genau!"
	incorrectMessage = "Nein :("
)

const (
	waitingForAnything = iota
	waitingForStart    = iota
	waitingForCheck    = iota
	waitingForAnswer   = iota
	waitingForAudio    = iota
	waitingForNoun     = iota
)

type UserJob struct {
	state              *UserState
	controls           *Contols
	chatID             int
	modeChoiceKeyboard tgbotapi.ReplyKeyboardMarkup
	keyboard           tgbotapi.ReplyKeyboardMarkup
	waitingState       int

	// training mode variables
	currentNouns   []Noun
	currentWordID  int
	numOfQuestions int
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

	job.modeChoiceKeyboard = tgbotapi.ReplyKeyboardMarkup{}
	buttons := [][]string{[]string{"/noun", "/audio"}, []string{"/verb", "/train"}}
	job.modeChoiceKeyboard.OneTimeKeyboard = true
	job.modeChoiceKeyboard.ResizeKeyboard = true
	job.modeChoiceKeyboard.Keyboard = buttons

	job.keyboard = tgbotapi.ReplyKeyboardMarkup{}
	buttons = [][]string{[]string{"der", "die"}, []string{"das", "die (pl)"}}
	job.keyboard.OneTimeKeyboard = true
	job.keyboard.ResizeKeyboard = true
	job.keyboard.Keyboard = buttons

	job.waitingState = waitingForAnything
}

func (job *UserJob) ProcessMessage(message string) {

	switch message {
	case "/help":
		job.SendMessage(helpMessage)
		return
	case "/train":
		job.SendMessage(trainingMessage)
		job.waitingState = waitingForStart
		return
	case "/noun":
		job.SendMessage(nounMessage)
		job.waitingState = waitingForNoun
		return
	case "/audio":
		job.SendMessage(audioMessage)
		job.waitingState = waitingForAudio
		return
	case "/verb":
		job.SendMessage(verbMessage)
		return
	}

	if job.waitingState == waitingForStart ||
		job.waitingState == waitingForCheck ||
		job.waitingState == waitingForAnswer {
		if message == "/stop" {
			job.SendMessage(stopMessage)
			job.waitingState = waitingForAnything
			return
		}
	}

	switch job.waitingState {
	case waitingForStart:
		if message == "/start" {
			job.SendNewWords()
		} else {
			job.SendMessage(helpMessage)
		}
	case waitingForCheck:
		if message == "/check" {
			job.SendSticker()
			job.SendQuestion()
		} else {
			job.SendMessage(readyMessage)
		}
	case waitingForAnswer:
		job.CheckAnswers(message)
	case waitingForAudio:
		job.SendAudio(message)
	case waitingForNoun:
		job.SendNoun(message)
	case waitingForAnything:
		job.SendMessage(helpMessage)
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
	job.waitingState = waitingForCheck
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
	message.ReplyMarkup = job.modeChoiceKeyboard
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
	job.waitingState = waitingForAnything
}

func (job *UserJob) SendNoun(message string) {
	job.waitingState = waitingForAnything
	infos := getNounInfo(message)
	if len(infos) <= 1 {
		job.SendMessage(noNounMessage)
		return
	}
	for _, info := range infos {
		job.SendMessage(info)
	}
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
