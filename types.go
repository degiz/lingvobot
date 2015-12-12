package main

import (
	"fmt"
	"github.com/Syfaro/telegram-bot-api"
)

type Noun struct {
	ID          int
	Article     string
	Noun        string
	Translation string
}

type UserState struct {
	channel chan tgbotapi.Message
	quit    chan int
	UserID  int
}

func (noun *Noun) getIDString() string {
	return fmt.Sprintf("noun:%d", noun.ID)
}

func getNounIDKey(id int) string {
	return fmt.Sprintf("noun:%d", id)
}

func (state *UserState) getIDString() string {
	return fmt.Sprintf("user:%d", state.UserID)
}

func getUserIDKey(id int) string {
	return fmt.Sprintf("user:%d", id)
}
