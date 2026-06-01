package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/FlameInTheDark/gochat/clients/bot/goclient"
)

var (
	token         string
	endpoint      string
	applicationID int64
	guildID       int64
	register      bool
)

func init() {
	flag.StringVar(&token, "t", os.Getenv("GOCHAT_BOT_TOKEN"), "GoChat bot token")
	flag.StringVar(&endpoint, "endpoint", "", "GoChat service URL")
	flag.Int64Var(&applicationID, "application", 0, "bot application/user ID")
	flag.Int64Var(&guildID, "guild", 0, "optional guild ID for guild commands")
	flag.BoolVar(&register, "register", false, "bulk overwrite example commands before opening the gateway")
	flag.Parse()
}

func main() {
	if token == "" {
		fmt.Println("missing bot token; pass -t or set GOCHAT_BOT_TOKEN")
		return
	}

	options := []goclient.Option{}
	if endpoint != "" {
		options = append(options, goclient.WithEndpoint(endpoint))
	}
	session, err := goclient.New(token, options...)
	if err != nil {
		fmt.Println("error creating GoChat session:", err)
		return
	}

	if register {
		if applicationID == 0 {
			fmt.Println("missing application ID; pass -application when -register is used")
			return
		}
		if err := registerCommands(context.Background(), session, applicationID, guildID); err != nil {
			fmt.Println("error registering commands:", err)
			return
		}
	}

	session.AddHandler(func(session *goclient.Session, interaction *goclient.Interaction) {
		data := interaction.ApplicationCommandData()
		fmt.Printf("interaction received: id=%d type=%d command=%q options=%d\n", interaction.ID, interaction.Type, data.Name, len(data.Options))
		switch interaction.Type {
		case goclient.InteractionAutocomplete:
			_ = session.InteractionRespond(context.Background(), interaction, &goclient.InteractionResponse{
				Type: goclient.InteractionApplicationCommandAutocompleteResult,
				Data: &goclient.InteractionResponseData{
					Choices: autocompleteChoices(data),
				},
			})
		case goclient.InteractionApplicationCommand:
			switch data.Name {
			case "hello":
				fmt.Println("hello action used")
				if err := session.InteractionRespond(context.Background(), interaction, &goclient.InteractionResponse{
					Type: goclient.InteractionResponseChannelMessageWithSource,
					Data: &goclient.InteractionResponseData{
						Content: "Hello from an application command.",
						Flags:   goclient.MessageFlagsEphemeral,
					},
				}); err != nil {
					fmt.Println("error responding to interaction:", err)
				}
			case "day":
				location := "your location"
				if option := data.Option("location"); option != nil && option.StringValue() != "" {
					location = option.StringValue()
				}
				if err := session.InteractionRespond(context.Background(), interaction, &goclient.InteractionResponse{
					Type: goclient.InteractionResponseChannelMessageWithSource,
					Data: &goclient.InteractionResponseData{
						Content: fmt.Sprintf("Forecast for %s: clear skies with a high chance of working slash commands.", location),
					},
				}); err != nil {
					fmt.Println("error responding to interaction:", err)
				}
			case "slow":
				_ = session.InteractionRespond(context.Background(), interaction, &goclient.InteractionResponse{
					Type: goclient.InteractionResponseDeferredChannelMessageSource,
				})
				go func() {
					time.Sleep(2 * time.Second)
					_, _ = session.InteractionResponseEdit(context.Background(), interaction.ApplicationID, interaction.Token, &goclient.InteractionResponseData{
						Content: "Deferred response finished.",
					})
					_, _ = session.FollowupMessageCreate(context.Background(), interaction.ApplicationID, interaction.Token, &goclient.InteractionResponseData{
						Content: "And here is a followup.",
					})
				}()
			}
		}
	})

	if err := session.Open(); err != nil {
		fmt.Println("error opening gateway:", err)
		return
	}
	defer func() { _ = session.Close() }()

	fmt.Println("application command bot is running; press Ctrl+C to exit")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}

func registerCommands(ctx context.Context, session *goclient.Session, applicationID, guildID int64) error {
	_, err := session.ApplicationCommandBulkOverwrite(ctx, applicationID, guildID, []*goclient.ApplicationCommand{
		{
			Type:        goclient.ApplicationCommandChatInput,
			Name:        "hello",
			Description: "Send an ephemeral greeting",
			Contexts:    []goclient.InteractionContextType{goclient.InteractionContextGuild, goclient.InteractionContextBotDM},
		},
		{
			Type:        goclient.ApplicationCommandChatInput,
			Name:        "slow",
			Description: "Demonstrate deferred responses and followups",
			Contexts:    []goclient.InteractionContextType{goclient.InteractionContextGuild, goclient.InteractionContextBotDM},
		},
		{
			Type:        goclient.ApplicationCommandChatInput,
			Name:        "day",
			Description: "Show today's forecast",
			Options: []*goclient.ApplicationCommandOption{
				{
					Type:        goclient.ApplicationCommandOptionString,
					Name:        "location",
					Description: "The name of the location",
					Required:    true,
				},
			},
			Contexts: []goclient.InteractionContextType{goclient.InteractionContextGuild, goclient.InteractionContextBotDM},
		},
		{
			Type:        goclient.ApplicationCommandChatInput,
			Name:        "search",
			Description: "Demonstrate autocomplete",
			Options: []*goclient.ApplicationCommandOption{
				{
					Type:         goclient.ApplicationCommandOptionString,
					Name:         "query",
					Description:  "Search text",
					Required:     true,
					Autocomplete: true,
				},
			},
			Contexts: []goclient.InteractionContextType{goclient.InteractionContextGuild, goclient.InteractionContextBotDM},
		},
	})
	return err
}

func autocompleteChoices(data goclient.ApplicationCommandInteractionData) []*goclient.ApplicationCommandOptionChoice {
	query := ""
	if option := data.FocusedOption(); option != nil {
		query = strings.ToLower(option.StringValue())
	}
	values := []string{"alpha", "beta", "release", "roadmap", "support"}
	choices := make([]*goclient.ApplicationCommandOptionChoice, 0, len(values))
	for _, value := range values {
		if query != "" && !strings.Contains(value, query) {
			continue
		}
		choices = append(choices, &goclient.ApplicationCommandOptionChoice{Name: value, Value: value})
	}
	return choices
}
