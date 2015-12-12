package main

import (
	"flag"
	"log"
)

func main() {

	var configFile string
	flag.StringVar(&configFile, "config", "config.json", "path to config file")
	flag.Parse()

	bot := &Bot{configFile: configFile}
	log.Printf("LingvoBot starting..")
	bot.Run()

}
