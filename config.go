package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	RedisAddress        string
	RedisPassword       string
	RedisDb             int64
	NounsPath           string
	TelegramBotTokenEnv string
	IvonaAccessKeyEnv   string
	IvonaAccessKeyToken string
	IvonaSecretKeyEnv   string
	IvonaSecretKeyToken string
	TelegramBotToken    string
}

// TODO: put redis config to json, only token shoud be in env
func getConfig(filepath string) (*Config, error) {

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	errDec := decoder.Decode(config)
	if errDec != nil {
		return nil, errDec
	}

	config.TelegramBotToken = os.Getenv(config.TelegramBotTokenEnv)
	config.IvonaAccessKeyToken = os.Getenv(config.IvonaAccessKeyEnv)
	config.IvonaSecretKeyToken = os.Getenv(config.IvonaSecretKeyEnv)

	return config, nil
}
