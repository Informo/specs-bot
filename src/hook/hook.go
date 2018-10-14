package hook

import (
	"strings"

	"matrix"
	"types"

	"gopkg.in/go-playground/webhooks.v5/github"
)

// HandlePullRequestPayload processes the payload of a pull request event received
// by the GitHub webhook. If the event's action is anothing other than
// "labeled", which means a label has been added to the pull request, it does
// nothing. If the action is "labeled", it loads the data of the SCS from the
// PR's data, and iterates through the PR's labels to get its type and SCSP
// state. If one of these two isn't defined, it does nothing more and returns.
// If both are filled, it sends a notice through the Matrix room. If the notice
// send fails, it returns with an error.
// If there's more than one SCS type and/or more than one SCSP state for a SCS,
// it returns and do nothing.
func HandlePullRequestPayload(
	pl github.PullRequestPayload, cli *matrix.Cli,
) (err error) {
	// Only process the "labeled" action.
	if pl.Action == "labeled" {
		pr := pl.PullRequest

		// Load the SCS data from the PR.
		data := types.SCSData{
			Number: pr.Number,
			Title:  pr.Title,
			URL:    pr.HTMLURL,
		}

		// Iterate through the PR's labels.
		for _, l := range pr.Labels {
			// All labels defined in the SCSP follow the form "xxx:yyy", such as
			// "xxx" is the type of information held by the label, and yyy is
			// that information.
			split := strings.Split(l.Name, ":")

			// "xxx" can either be "type", which is the type of the changes
			// submitted (typo, behaviour), or "scsp", which is the SCS's SCSP
			// state.
			switch split[0] {
			case "type":
				// If more than one type is defined, return and do nothing.
				if len(data.Type) > 0 {
					return
				}
				data.Type = split[1]
			case "scsp":
				// If more than one SCSP state is defined, return and do noting.
				if len(data.State) > 0 {
					return
				}
				data.State = split[1]
			}
		}

		// Check if both the type and SCSP state are defined.
		if len(data.Type) == 0 || len(data.State) == 0 {
			return
		}

		// Emit a notice to the configured Matrix room.
		err = cli.SendNotice(data)
	}

	return
}
