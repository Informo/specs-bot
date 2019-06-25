package main

import (
	"flag"
	"net/http"

	"github.com/Informo/specs-bot/config"
	"github.com/Informo/specs-bot/database"
	"github.com/Informo/specs-bot/hook"
	"github.com/Informo/specs-bot/matrix"

	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/webhooks.v5/github"
)

var (
	configFile = flag.String("config", "config.yaml", "Path for the configuration file")
	debug      = flag.Bool("debug", false, "Enable debug logging")
)

func main() {
	// Parse the command line flags.
	flag.Parse()

	// Configure logrus.
	logConfig()

	// Load the configuration.
	cfg, err := config.Load(*configFile)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Debug("Configuration loaded")

	// Instantiate a Matrix client.
	cli, err := matrix.NewCli(
		cfg.Matrix.HSURL, cfg.Matrix.MXID, cfg.Matrix.AccessToken, cfg,
	)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Debug("Matrix client instantiated")

	// Instantiate the database and prepare statements.
	db, err := database.NewDatabase(cfg)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Debug("Database instantiated")

	// Instantiate a GitHub webhook.
	h, err := github.New(github.Options.Secret(cfg.Webhook.Secret))
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Debug("GitHub webhook instantiated")

	// Define the HTTP handler for the webhook.
	http.HandleFunc(cfg.Webhook.Path, func(w http.ResponseWriter, r *http.Request) {
		var payload interface{}
		// Retrieve the payload if the event is a pull request event.
		payload, err = h.Parse(r, github.PullRequestEvent, github.IssuesEvent)
		if err != nil {
			// If the event isn't a pull request event, notify the sender that
			// the request isn't within what's expected and return.
			if err == github.ErrEventNotFound {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			logrus.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Handle both issues and pull requests payloads.
		switch payload.(type) {
		case github.PullRequestPayload:
			err = hook.HandlePullRequestPayload(
				payload.(github.PullRequestPayload), cli, db,
			)
			break
		case github.IssuesPayload:
			err = hook.HandleIssuesPayload(
				payload.(github.IssuesPayload), cli, db,
			)
			break
		}

		// If any of the handler or workflow returned with an error, log it and
		// tell the user something went wrong.
		if err != nil {
			logrus.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	logrus.WithField("path", cfg.Webhook.Path).Debug("Defined HTTP handler")

	// Start the HTTP server.
	logrus.WithField("listen_addr", cfg.Webhook.ListenAddr).Info("Starting web server")
	if err = http.ListenAndServe(cfg.Webhook.ListenAddr, nil); err != nil {
		logrus.Panic(err)
	}
}
