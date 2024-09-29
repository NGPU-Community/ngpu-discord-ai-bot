package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/nGPU/bot/header"
)

func makeContent(before string, cmdName string, content string) string {
	tmpContent := ""
	tmpContent += fmt.Sprintf("%s\n", before)
	tmpContent += fmt.Sprintf("**%s**\n", cmdName)
	if content != "" {
		tmpContent += fmt.Sprintf("```%s```", content)
	}
	return tmpContent
}

func makeData(content string, title, description string, color int) *discordgo.InteractionResponseData {
	data := &discordgo.InteractionResponseData{
		Content: content,
		Embeds: []*discordgo.MessageEmbed{&discordgo.MessageEmbed{
			Title:       title,
			Description: description,
			Color:       color,
		}},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					header.HomeButton,
				},
			},
		},
	}
	return data
}
