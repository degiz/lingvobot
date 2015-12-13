#Lingvobot

A bot for Telegram messenger that helps you to learn some foreign words.
Implemented in go, packed with a Dockerfile to build yet another container.

For now, it's designed to help you with german nouns and articles.

Lingvobot can show you an article, plural form and all the casuses of a noun. For now it uses [wiktionary](https://de.wiktionary.org) for that.

Lingvobot shows you some nouns with a translation and an article, then asks you questions with a custom "german articles" keyboard.

Also the Lingvobot can send you pronunciation of any text as an audio file, for now it uses [ivona](https://www.ivona.com) for that.

The Makefile expects you to set up:
* TELEGRAM_BOT_TOKEN environment variable with an Telegram bot API token.
* IVONA_ACCESS_KEY environment variable with an IVONA access key
* IVONA_SECRET_KEY environment variable with an IVONA secret key

The LingvoBot is deployed and active since I use it myself, drop a message to @NerdLingvoBot
Please note, that I don't give any uptime garanties.

Goodluck learning foreign languages!

![Screenshot of bot-in-action](/images/screenshot1.png?raw=true "Screenshot of bot-in-action")
![Screenshot of bot-in-action](/images/screenshot2.png?raw=true "Screenshot of bot-in-action")