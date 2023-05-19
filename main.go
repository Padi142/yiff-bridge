package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	fmt.Println("Priming Yiff-bridge")
	fmt.Println("Prepare for some quality furry photos")
	telegramChannel := make(chan string)
	discordChannel := make(chan string)

	config, err := getConfig()
	if err != nil {
		return
	}

	go telegram(telegramChannel, discordChannel, config.TelegramToken, config.TelegramChannel)

	go discord(telegramChannel, discordChannel, config.DiscordToken, config.DiscordChannel)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func discord(telegramChannel, discordChannel chan string, discordId string, discordChannelId string) {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + discordId)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// We need information about guilds (which includes their channels),
	// messages and voice states.
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	fmt.Println("Discord up")

	go func() {
		for {
			command := <-discordChannel
			if strings.HasPrefix(command, "yiff_detected:") {
				yiff := strings.Replace(command, "yiff_detected:", "", -1)
				_, err = dg.ChannelMessageSend(discordChannelId, yiff)
				if err != nil {
					fmt.Println("Error sending message:", err)
				}
			}

		}

	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}
func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateGameStatus(0, "yiff-gardener")

}

func telegram(telegramChannel, discordChannel chan string, telegramId string, telegramChannelId int64) {
	bot, err := tgbotapi.NewBotAPI(telegramId)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Telegram up")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {

		if update.ChannelPost != nil {
			if update.ChannelPost.Chat.LinkedChatID != telegramChannelId {
				return
			}
			if len(update.ChannelPost.Photo) == 0 {
				if update.ChannelPost.Document == nil {
					discordChannel <- "yiff_detected:" + update.ChannelPost.Text
					continue
				}
				config := tgbotapi.FileConfig{FileID: update.ChannelPost.Document.FileID}
				downloader, _ := bot.GetFile(config)
				link := downloader.Link(telegramId)
				log.Printf("Yiff detected")

				discordChannel <- "yiff_detected:" + link

				continue
			}
			config := tgbotapi.FileConfig{FileID: update.ChannelPost.Photo[len(update.ChannelPost.Photo)-1].FileID}
			downloader, _ := bot.GetFile(config)
			link := downloader.Link(telegramId)
			log.Printf("Yiff detected")

			discordChannel <- "yiff_detected:" + link

		}
	}
}

type Config struct {
	// defining struct variables
	DiscordToken    string `json:"discord_token"`
	TelegramToken   string `json:"telegram_token"`
	DiscordChannel  string `json:"discord_channel_id"`
	TelegramChannel int64  `json:"telegram_channel"`
}

func getConfig() (Config, error) {
	var config Config
	configFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println("Error opening config.json. Make sure it exists")
		return config, nil
	}
	byteValue, _ := io.ReadAll(configFile)

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		fmt.Println("Error parsing config.json")
		return config, nil
	}

	defer configFile.Close()

	return config, nil
}
