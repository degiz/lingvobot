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
	err_dec := decoder.Decode(config)
	if err_dec != nil {
		return nil, err_dec
	}

	config.TelegramBotToken = os.Getenv(config.TelegramBotTokenEnv)

	return config, nil
}
