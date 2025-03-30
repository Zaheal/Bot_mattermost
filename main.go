package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"time"

	"github.com/joho/godotenv"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/rs/zerolog"
)

func init() {
	
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	app := &application{
		logger: zerolog.New(
			zerolog.ConsoleWriter{
				Out: os.Stdout,
				TimeFormat: time.RFC822,
			},
		).With().Timestamp().Logger(),
	}

	app.config = getConfig()
	app.logger.Info().Str("config", fmt.Sprint(app.config)).Msg("")

	setupGracefulShutdown(app)

	app.mmClient = model.NewAPIv4Client(app.config.MATTERMOST_URL)

	app.mmClient.SetToken(app.config.ACCESS_TOKEN)

	if user, resp, err := app.mmClient.GetUser("me", ""); err != nil {
		app.logger.Fatal().Err(err).Msg("Could not login")
	} else {
		app.logger.Debug().Interface("user", user).Interface("resp", resp).Msg("")
		app.logger.Info().Msg("Logged in to mm")
		app.mmUser = user
	}

	if team, resp, err := app.mmClient.GetTeamByName(app.config.TEAM_NAME, ""); err != nil {
		app.logger.Fatal().Err(err).Msg("Could not find team. Is this bot a member?")
	} else {
		app.logger.Debug().Interface("team", team).Interface("resp", resp).Msg("")
		app.mmTeam = team
	}

	if channel, resp, err := app.mmClient.GetChannel(
		app.config.CHANNEL_ID, "",
	); err != nil {
		app.logger.Fatal().Err(err).Msg("Could not find channel. Is this bot added to that channel?")
	} else {
		app.logger.Debug().Interface("channel", channel).Interface("resp", resp).Msg("")
		app.mmChannel = channel
	}
	app.initTarantool()

	listenToEvents(app)
}


func setupGracefulShutdown(app *application) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			if app.mmWebSocket != nil {
				app.logger.Info().Msg("Closing websocket connection")
				app.mmWebSocket.Close()
			}
			app.logger.Info().Msg("Shutting down")
			os.Exit(0)
		}
	}()
}

func listenToEvents(app *application) {
	var err error
	failCount := 0

	for {
		app.mmWebSocket, err = model.NewWebSocketClient4(
			fmt.Sprintf("ws://%s", app.config.MATTERMOST_URL[7:]),
			app.mmClient.AuthToken,
		)
		if err != nil {
			app.logger.Warn().Err(err).Msg("MM websocket disconnect, retrying")
			failCount += 1
			continue
		}
		app.logger.Info().Msg("Websocket connected")

		app.mmWebSocket.Listen()

		for event := range app.mmWebSocket.EventChannel {
			go handleWebSocketEvent(app, event)
		}
	}
}

func handleWebSocketEvent(app *application, event *model.WebSocketEvent) {
	
	if event.GetBroadcast().ChannelId != app.mmChannel.Id {
		return
	}

	if event.EventType() != model.WebsocketEventPosted {
		return
	}

	post := &model.Post{}
	err := json.Unmarshal([]byte(event.GetData()["post"].(string)), &post)
	if err != nil {
		app.logger.Error().Err(err).Msg("Could not cast event to *model.Post")
	}

	if post.UserId == app.mmUser.Id {
		return
	}

	handlePost(app, post)
}

func handlePost(app *application, post *model.Post) {
	app.logger.Debug().Str("message", post.Message).Msg("")
	app.logger.Debug().Interface("post", post).Msg("")

	if matched, _ := regexp.MatchString(`^/vote\b`, post.Message); matched {
		app.handlerVoteCommands(post)
		return
	} else {
		sendMsg(app, "I can`t understand", post.Id)
	}
}

func sendMsg(app *application, msg string, channelId string) {

	post := &model.Post{}
	post.ChannelId = app.mmChannel.Id
	post.Message = msg

	post.ChannelId = channelId

	if _, _, err := app.mmClient.CreatePost(post); err != nil {
		app.logger.Error().Err(err).Str("ChannelID", channelId).Msg("Failed to create post")
	}
}