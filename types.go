package main

import (
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
)

type Noun struct {
	Id          int
	Article     string
	Noun        string
	Translation string
}

type UserState struct {
	channel chan tgbotapi.Message
	quit    chan int
	UserId  int
}

func (self *Noun) getIdString() string {
	return fmt.Sprintf("noun:%d", self.Id)
}

func getNounIdKey(id int) string {
	return fmt.Sprintf("noun:%d", id)
}

func (self *UserState) getIdString() string {
	return fmt.Sprintf("user:%d", self.UserId)
}

func getUserIdKey(id int) string {
	return fmt.Sprintf("user:%d", id)
}
