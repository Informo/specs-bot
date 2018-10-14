package main

import (
	"flag"
	"fmt"
	"net/http"

	"config"
	"hook"
	"matrix"

	"gopkg.in/go-playground/webhooks.v5/github"
)

var (
	configFile = flag.String("config", "config.yaml", "Path for the configuration file")
)

func main() {
	// Parse the command line flags.
	flag.Parse()

	// Load the configuration.
	cfg, err := config.Load(*configFile)
	if err != nil {
		panic(err)
	}

	// Instantiate a Matrix client.
	cli, err := matrix.NewCli(
		cfg.Matrix.HSURL, cfg.Matrix.MXID, cfg.Matrix.AccessToken, cfg,
	)
	if err != nil {
		panic(err)
	}

	// Instantiate a GitHub webhook.
	h, err := github.New(github.Options.Secret(cfg.Webhook.Secret))
	if err != nil {
		panic(err)
	}

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

			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch payload.(type) {
		case github.PullRequestPayload:
			err = hook.HandlePullRequestPayload(
				payload.(github.PullRequestPayload), cli,
			)
			break
		case github.IssuesPayload:
			err = hook.HandleIssuesPayload(
				payload.(github.IssuesPayload), cli,
			)
			break
		}

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	// Start the HTTP server.
	if err = http.ListenAndServe(cfg.Webhook.ListenAddr, nil); err != nil {
		panic(err)
	}
}
