package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"

	"github.com/shaharia-lab/slackbot"
)

func main() {
	// Load configuration
	config, err := slackbot.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	bot, err := slackbot.NewBot(config, logger)
	if err != nil {
		logger.Error("Failed to create bot: %v", err)
		return
	}

	// Set up HTTP handler for Slack events
	http.HandleFunc("/slack/events", func(w http.ResponseWriter, r *http.Request) {
		logger.Info(fmt.Sprintf("Received request: %s %s", r.Method, r.URL.Path))

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify the request is coming from Slack
		sv, err := slack.NewSecretsVerifier(r.Header, config.SlackSigninSecret)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if _, err := sv.Write(body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := sv.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(
			json.RawMessage(body),
			slackevents.OptionNoVerifyToken(),
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Handle URL verification challenge
		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(r.Challenge))
			return
		}

		// Process the event
		if err := bot.HandleEvent(eventsAPIEvent); err != nil {
			log.Printf("Error handling event: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	})

	// Start HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
