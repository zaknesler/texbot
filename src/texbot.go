// Author: Zak Nesler
// Date: 2022-05-07
//
// This file is a part of TexBot. TexBot is a Discord bot that waits for
// messages containing LaTeX expressions and parses, renders, and sends the
// result as an image.
//
// Requires a valid Discord bot token with proper permissions (see README.md).

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

// Discord API bot token
var TOKEN string

func init() {
	log.SetPrefix("texbot: ")

	// If the token is provided as an environment variable, use it
	if os.Getenv("DISCORD_TOKEN") != "" {
		TOKEN = os.Getenv("DISCORD_TOKEN")
		return
	}

	// Otherwise, expect a required command line flag
	flag.StringVar(&TOKEN, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	// Kill the program if no token was provided
	if len(TOKEN) == 0 {
		log.Fatalln("No token provided!\nUsage: texbot -t <token>")
	}

	// Create a new Discord bot session
	s, err := discordgo.New("Bot " + TOKEN)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	// Register the ready and message handlers
	s.AddHandler(onReady)
	s.AddHandler(onMessageCreate)

	// Open the connection
	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	// Close the connection after signal interrupt
	defer s.Close()

	// Wait for a termination signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
}

// onReady is called when the bot has come online and is ready to go
func onReady(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)

	// Set listening status of the bot to helpful example usage message
	err := s.UpdateListeningStatus("$$ <LaTeX> $$")
	if err != nil {
		log.Fatalf("Could not set listening status: %v", err)
	}
}

// onMessageCreate is called when a message is created
func onMessageCreate(sess *discordgo.Session, msg *discordgo.MessageCreate) {
	// Ignore all messages created by the bot
	if msg.Author.ID == sess.State.User.ID {
		return
	}

	// Check if message contains LaTeX expression
	match := ParseString(msg.Content)

	// If not, don't continue
	if !match.HasMatch {
		return
	}

	log.Printf("LaTeX detected in message: %v", msg.Message.Content)

	// Asynchronously handle the match (parse, render, send)
	go handleMatch(match, sess, msg)
}

// Handle operating on a LaTeX expression that was detected in a message
func handleMatch(match ParsedString, sess *discordgo.Session, msg *discordgo.MessageCreate) {
	// Render the LaTeX expression
	result := Render(match)

	// If there was an error rendering the expression, send it to the user
	if result.Err != nil {
		log.Printf("Couldn't render: %v", result.Err)

		// Send the error message as a reply to the original message
		_, err := sess.ChannelMessageSendReply(msg.ChannelID, fmt.Sprintf("Couldn't render expression: `%s` ```%s```", match.Expr, result.ParseErrMsg), msg.Reference())
		if err != nil {
			log.Printf("Couldn't send reply: %v", err)
		}

		return
	}

	// Defer file closing (this only removes from memory, the actual file is
	// deleted after rendering)
	defer result.File.Close()

	var content string = ""
	var ref *discordgo.MessageReference = nil

	// If the message was sent as a direct message to the Bot account, send the
	// image as a reply to the original message as a bot cannot delete messages
	// sent as a DM
	if msg.GuildID == "" {
		ref = msg.Reference()
	} else {
		// Mention the user who sent the message and include the original expression
		content = fmt.Sprintf("%s `%s`", msg.Author.Mention(), match.Expr)

		// Delete original message as the bot will send its own message with
		// the original expression
		err := sess.ChannelMessageDelete(msg.ChannelID, msg.ID)
		if err != nil {
			log.Printf("Couldn't delete message: %v", err)
		}
	}

	var channel string = msg.ChannelID

	// If the message was sent in a thread, send the message in the thread and
	// not in the thread's parent channel
	if msg.Thread != nil {
		channel = msg.Thread.ID
	}

	// Send the image in a message as an attached file
	_, err := sess.ChannelMessageSendComplex(channel, &discordgo.MessageSend{
		Content:   content,
		Reference: ref,
		Files: []*discordgo.File{{
			Name:        "render.png",
			ContentType: "image/png",
			Reader:      result.File,
		}},
	})
	if err != nil {
		log.Fatalf("Couldn't send message: %v", err)
	}
}
