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

		labels := []string{}
		for _, l := range pr.Labels {
			labels = append(labels, l.Name)
		}

		err = handleSubmission(pr.Number, pr.Title, pr.HTMLURL, labels)
	}

	return
}

func HandleIssuesPayload(
	pl github.IssuesPayload, cli *matrix.Cli,
) (err error) {
	// Only process the "labeled" action.
	if pl.Action == "labeled" {
		issue := pl.Issue

		labels := []string{}
		for _, l := range issue.Labels {
			labels = append(labels, l.Name)
		}

		err = handleSubmission(issue.Number, issue.Title, issue.HTMLURL, labels)
	}

	return
}

func handleSubmission(
	number int64, title string, url string, labels []string,
) (err error) {
	data := types.SCSData{
		Number: number,
		Title:  title,
		URL:    url,
	}

	unsplittableLabels := []string{}

	for _, l := range labels {
		// All labels defined in the SCSP follow the form "xxx:yyy", such as
		// "xxx" is the type of information held by the label, and yyy is
		// that information.
		split := strings.Split(l, ":")

		if len(split) < 2 {
			unsplittableLabels = append(unsplittableLabels, l)
			continue
		}

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
			break
		case "scsp":
			// If more than one SCSP state is defined, return and do noting.
			if len(data.State) > 0 {
				return
			}
			data.State = split[1]
			break
		}
	}

	if len(data.Type) == 0 || len(data.State) == 0 {
		return cli.SendNoticeWithUnsplitLabels(data, unsplittableLabels)
	}

	// Emit a notice to the configured Matrix room.
	return cli.SendNoticeWithTypeAndState(data)
}
