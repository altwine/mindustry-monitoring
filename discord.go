package main

import (
	"bytes"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/fogleman/gg"
)

var dg *discordgo.Session

func initDiscordBot() {
	var err error
	dg, err = discordgo.New("Bot " + DISCORD_TOKEN)
	if err != nil {
		log.Fatal("error creating session: ", err)
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	dg.AddHandler(messageHandler)

	err = dg.Open()
	if err != nil {
		log.Fatal("error opening connection: ", err)
	}
}

func destroyDiscordBot() {
	dg.Close()
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	messageSplitted := strings.Split(m.Content, " ")
	if len(messageSplitted) == 1 {
		if messageSplitted[0] == "!is-working" {
			s.ChannelMessageSend(m.ChannelID, "Yes")
		}
		return
	}
	if len(messageSplitted) == 2 || len(messageSplitted) == 3 {
		if messageSplitted[0] == "!info" {
			serverAddress := messageSplitted[1]
			if len(serverAddress) == 0 {
				s.ChannelMessageSend(m.ChannelID, "Invalid server address format (try IPv4 or hostname)")
				return
			}

			serverPort := "6567"
			if len(messageSplitted) == 3 {
				serverPort = messageSplitted[2]
			}

			serverHost := serverAddress + ":" + serverPort

			stats, err := getStatsByAddress(serverHost, 12)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "cant find your server in our database")
				return
			}

			serverInfo, ok := infoObjects[serverHost]
			if !ok {
				s.ChannelMessageSend(m.ChannelID, "cant find info for your server in our database")
				return
			}

			dc := gg.NewContext(width, height)
			genImage(dc, *serverInfo, stats)
			var buf bytes.Buffer
			if err := dc.EncodePNG(&buf); err != nil {
				s.ChannelMessageSend(m.ChannelID, "cant gen image")
				return
			}
			s.ChannelFileSend(m.ChannelID, "image.png", &buf)
		}
		return
	}
}
