package main

import (
	"flag"
	"log"
)

func main() {

	var config_file string
	flag.StringVar(&config_file, "config", "config.json", "path to config file")
	flag.Parse()

	bot := &Bot{config_file: config_file}
	log.Printf("LingvoBot starting..")
	bot.Run()

}
