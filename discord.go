package main

import (
	"bytes"
	"log"

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

	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		cmd := &discordgo.ApplicationCommand{
			Name:        "info",
			Description: "Show server info",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "address",
					Description: "Server IPv4 address or hostname",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "port",
					Description: "Server port (default: 6567)",
					Required:    false,
				},
			},
		}
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("Cannot create command: %v", err)
		} else {
			log.Println("Command /info registered")
		}
	})

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			data := i.ApplicationCommandData()
			if data.Name == "info" {
				var address, port string
				for _, opt := range data.Options {
					switch opt.Name {
					case "address":
						address = opt.StringValue()
					case "port":
						port = opt.StringValue()
					}
				}
				if port == "" {
					port = "6567"
				}

				host := address + ":" + port

				stats, err := getStatsByAddress(host, 12)
				if err != nil {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "cant find your server in our database",
						},
					})
					return
				}

				serverInfo, ok := infoObjects[host]
				if !ok {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "cant find info for your server in our database",
						},
					})
					return
				}

				dc := gg.NewContext(width, height)
				genImage(dc, *serverInfo, stats)
				var buf bytes.Buffer
				if err := dc.EncodePNG(&buf); err != nil {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "cant gen image",
						},
					})
					return
				}

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Files: []*discordgo.File{
							&discordgo.File{
								Name:   "image.png",
								Reader: &buf,
							},
						},
					},
				})
			}
		}
	})

	err = dg.Open()
	if err != nil {
		log.Fatal("error opening connection: ", err)
	}
}

func destroyDiscordBot() {
	dg.Close()
}
