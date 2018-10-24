package hook

import (
	"strings"

	"matrix"
	"types"

	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/webhooks.v5/github"
)

// HandlePullRequestPayload processes the payload of a pull request event
// received by the GitHub webhook. If the event's action is related to labels
// (i.e. "(un)labeled"), it extracts the PR's labels' names and calls
// handleSubmission with the list of names and some specific data regarding the
// PR, which will then process the extracted data and trigger the generation
// and sending of a notice to the Matrix rooms.
// Returns and do nothing if the event's action isn't related to labels, or if
// handleSubmission (or subsequent function calls) decided there was no need to
// send a notice out.
// Returns with an error if handleSubmission or any subsequent function call
// returned with an error.
func HandlePullRequestPayload(
	pl github.PullRequestPayload, cli *matrix.Cli,
) (err error) {
	logrus.WithField("action", pl.Action).Debug("Got PR event payload")

	// Only process the label-related action.
	if pl.Action == "labeled" || pl.Action == "unlabeled" {
		pr := pl.PullRequest

		// Retrieve the labels' names.
		labels := []string{}
		for _, l := range pr.Labels {
			labels = append(labels, l.Name)
		}

		err = handleSubmission(pr.Number, pr.Title, pr.HTMLURL, labels, cli)
	}

	return
}

// HandleIssuesPayload processes the payload of an issue event received by the
// GitHub webhook. If the event's action is related to labels (i.e.
// "(un)labeled"), it extracts the issue's labels' names and calls
// handleSubmission with the list of names and some specific data regarding the
// issue, which will then process the extracted data and trigger the generation
// and sending of a notice to the Matrix rooms.
// Returns and do nothing if the event's action isn't related to labels, or if
// handleSubmission (or subsequent function calls) decided there was no need to
// send a notice out.
// Returns with an error if handleSubmission or any subsequent function call
// returned with an error.
func HandleIssuesPayload(
	pl github.IssuesPayload, cli *matrix.Cli,
) (err error) {
	logrus.WithField("action", pl.Action).Debug("Got issue event payload")

	// Only process the label-related actions.
	if pl.Action == "labeled" || pl.Action == "unlabeled" {
		issue := pl.Issue

		// Retrieve the labels' names.
		labels := []string{}
		for _, l := range issue.Labels {
			labels = append(labels, l.Name)
		}

		err = handleSubmission(
			issue.Number, issue.Title, issue.HTMLURL, labels, cli,
		)
	}

	return
}

// handleSubmission uses the given data referring to a submission to decide
// which workflow to use for the generation and sending of a Matrix notice for
// this submission update. It implements bot the Informo SCSP
// (https://specs.informo.network/introduction/scsp/) and a generic workflow
// which should work with most GitHub-driven submission workflow.
// Return and do nothing if there's too much information (i.e. more than one
// matching label name) for the submission's type or SCSP state, as we don't
// know what to do in this case (and the safer way to handle it is to do
// nothing).
// Returns with an error if either the Informo specific workflow or the generic
// one returns with an error.
func handleSubmission(
	number int64, title string, url string, labels []string, cli *matrix.Cli,
) (err error) {
	logDebugEntry := logrus.WithFields(logrus.Fields{
		"number": number,
		"title":  title,
		"url":    url,
		"labels": labels,
	})

	logDebugEntry.Debug("Handling submission")

	data := types.SCSData{
		Number: number,
		Title:  title,
		URL:    url,
	}

	unsplittableLabels := []string{}

	var l string
	for _, l = range labels {
		// All labels defined in the Informo SCSP follow the form "xxx:yyy",
		// such as "xxx" is the type of information held by the label, and yyy
		// is that information.
		split := strings.Split(l, ":")

		// If the label name couldn't be split, store it in a slice that will be
		// used if either the submission's type or state couldn't be determined,
		// and skip to the next iteration (as there's not enough data to
		// determine a specific type or state from this label name).
		if len(split) < 2 {
			logDebugEntry.WithField("name", l).Debug("Label name couldn't be split")
			unsplittableLabels = append(unsplittableLabels, l)
			continue
		}

		// Implementation of Informo's SCSP: extract the submission's type or
		// SCSP state from the label name.
		// "xxx" can either be "type", which is the type of the changes
		// submitted (typo, behaviour), or "scsp", which is the SCS's SCSP
		// state.
		switch split[0] {
		case "type":
			// If more than one type is defined, return and do nothing.
			if len(data.Type) > 0 {
				logDebugEntry.WithField("type", split[1]).Debug("Got another type, aborting")
				return
			}
			data.Type = split[1]
			logDebugEntry.WithField("type", data.Type).Debug("Got the submission type")
		case "scsp":
			// If more than one SCSP state is defined, return and do noting.
			if len(data.State) > 0 {
				logDebugEntry.WithField("state", split[1]).Debug("Got another state, aborting")
				return
			}
			data.State = split[1]
			logDebugEntry.WithField("state", data.State).Debug("Got the SCSP state")
		default:
			// If the first element in the split doesn't follow the Informo
			// SCSP, it means we should process this label name with the generic
			// workflow if we can (i.e. if a type or state can't be extracted
			// from other label names).
			logDebugEntry.WithField("name", l).Debug("Label name could be split but doesn't implement the SCSP")
			unsplittableLabels = append(unsplittableLabels, l)
			break
		}
	}

	// Redefine the log entry's fields to append the type and state now that we
	// have both of them in their definite state (i.e. their finite value or we
	// know one or more haven't been provided).
	logDebugEntry = logDebugEntry.WithFields(logrus.Fields{
		"type":  data.Type,
		"state": data.State,
	})

	// If the submission's type or SCSP state couldn't be determined from the
	// label names, it can either mean that the submission doesn't implement the
	// Informo SCSP, or a list of labels implementing it will come in a future
	// payload. To process the former case and ignore the latter, we use a
	// generic workflow that only processes labels that couldn't be split
	// accordingly with the Informo SCSP and for which a message string has been
	// defined.
	if len(data.Type) == 0 || len(data.State) == 0 {
		logDebugEntry.Debug("Calling the generic workflow")
		return cli.SendNoticeWithUnsplitLabels(data, unsplittableLabels)
	}

	// At this point we're pretty sure the submission implements Informo's SCSP,
	// so we use the dedicated workflow.
	logDebugEntry.Debug("Calling the Informo SCSP dedicated workflow")
	return cli.SendNoticeWithTypeAndState(data)
}
