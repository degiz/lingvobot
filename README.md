#Lingvobot

A bot for Telegram messenger that helps you to learn some foreign words.
Implemented in go, packed with a Dockerfile to build yet another container.

For now, it's designed to help you with german nouns and articles, look into nouns.txt for examples.
Lingvobot shows you some nouns with a translation and an article, then asks you questions with a custom "german articles" keyboard.

The Makefile expects you to set up TELEGRAM_BOT_TOKEN environment variable with an Telegram bot API token.

Goodluck learning foreign languages!

![Screenshot of bot-in-action](/images/screenshot.png?raw=true "Screenshot of bot-in-action")