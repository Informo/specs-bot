package matrix

import (
	"strings"
	"text/template"

	"github.com/Informo/specs-bot/config"
	"github.com/Informo/specs-bot/types"

	"github.com/matrix-org/gomatrix"
	"github.com/sirupsen/logrus"
)

// Cli is a representation of a Matrix client, containing both the gomatrix
// client and the configuration.
type Cli struct {
	c        *gomatrix.Client
	cfg      *config.Config
	prevMsgs map[int64]string
}

// NewCli creates and returns an instance of the Cli structure from the Matrix
// connection information and the configuration provided.
// Returns an error if the gomatrix client failed to initialise.
func NewCli(
	hsURL string, mxid string, accessToken string, cfg *config.Config,
) (cli *Cli, err error) {
	cli = new(Cli)
	cli.cfg = cfg
	cli.c, err = gomatrix.NewClient(hsURL, mxid, accessToken)
	cli.prevMsgs = make(map[int64]string)
	return
}

// SendNoticeWithTypeAndState generates a notice message from the SCS data and
// then sends the said message as a notice to the configured Matrix rooms.
// Returns an error if the message could not be generated or if the notice could
// not be sent to the Matrix rooms.
// Returns and do nothing if there's no message string available for the SCS's
// SCSP state.
func (c *Cli) SendNoticeWithTypeAndState(data *types.SCSData) (err error) {
	logDebugEntry := logrus.WithFields(logrus.Fields{
		"number": data.Number,
		"title":  data.Title,
		"url":    data.URL,
		"type":   data.Type,
		"state":  data.State,
	})

	// Check if there's a message string available for the given SCS type and
	// SCSP state.
	var ok bool
	data.Message, ok = c.cfg.Notices.Strings[data.Type][data.State]
	if !ok {
		logDebugEntry.Debug("Could not find a type-specific message string, searching into global message strings")
		// If a message string could not be found for the given SCS type and
		// SCSP state, check if there's a type-independant message string for
		// this SCSP state.
		data.Message, ok = c.cfg.Notices.Strings["global"][data.State]

		// If no string could be found, return and do nothing.
		if !ok {
			logDebugEntry.Debug("Could not find a global message string for the given state")
			return
		}

		logDebugEntry.Debug("Got a global message string")
	} else {
		logDebugEntry.Debug("Got a type-specific message string")
	}

	return c.sendNotice(data)
}

// SendNoticeWithUnsplitLabels generates a notice message from the pull
// request's or issue's labels that couldn't be split accordingly with Informo's
// SCSP. It then sends the message to the configured Matrix room. It is meant to
// be used as a fallback if either the type or the state couldn't be determined
// (i.e. if the proposal doesn't implement Informo's SCSP).
// Returns with an error if the notice message could not be generated or sent.
// Returns and do nothing if there's no message string matching any of the given
// labels, or if there was more than one match.
func (c *Cli) SendNoticeWithUnsplitLabels(
	data *types.SCSData, unsplitLabels []string,
) (err error) {
	logDebugEntry := logrus.WithFields(logrus.Fields{
		"number": data.Number,
		"title":  data.Title,
		"url":    data.URL,
		"labels": unsplitLabels,
	})

	messages := make([]string, 0)

	var ok bool
	var match string
	for _, l := range unsplitLabels {
		// If we have another match when there's already a message loaded in,
		// we don't know what message to use. In this case, don't do anything.
		if match, ok = c.cfg.Notices.Strings["global"][l]; ok && len(data.Message) > 0 {
			logDebugEntry.WithField("name", l).Debug("Found another message string for label name, aborting")
			return
		} else if ok {
			messages = append(messages, match)
			logDebugEntry.WithField("name", l).Debug("Found a message string for label name")
		}
	}

	switch len(messages) {
	case 0:
		// No message could be found for any of the labels.
		logDebugEntry.Debug("Could not find a message for any label name")
		return
	case 1:
		// Only 1 message has been found, we don't need to do any copy.
		data.Message = messages[0]
		return c.sendNotice(data)
	default:
		// More than 1 message has been found found, we copy the structure as much
		// as necessary with the different messages then send them.
		var msg string
		for _, msg = range messages {
			if err = c.sendNotice(data.CopyWithMsg(msg)); err != nil {
				return
			}
		}
	}

	return nil
}

// sendNotice uses the given data to generate the full notice message for this
// submission update from the configured template, and send it to the Matrix
// rooms.
// Returns and do nothing if the latest message sent for this submission is the
// same as the message for this update.
// Returns with an error it there was an issue generating the notice message
// from the configured template, or sending it out as a notice to the Matrix
// room.
func (c *Cli) sendNotice(data *types.SCSData) (err error) {
	logEntry := logrus.WithFields(logrus.Fields{
		"number":  data.Number,
		"title":   data.Title,
		"url":     data.URL,
		"message": data.Message,
		"type":    data.Type,
		"state":   data.State,
	})

	if msg, ok := c.prevMsgs[data.Number]; ok && strings.Compare(msg, data.Message) == 0 {
		logEntry.Debug("Already sent this update for this submission")
		return
	}

	c.prevMsgs[data.Number] = data.Message

	// Load the template defined in the configuration file. The "message" name
	// used here is not important.
	tmpl, err := template.New("message").Parse(c.cfg.Notices.Pattern)
	if err != nil {
		logEntry.Debug("Could not load template")
		return
	}

	// Generate the notice message from the configured template and the SCS's
	// data.
	var b strings.Builder
	if err = tmpl.Execute(&b, data); err != nil {
		logEntry.Debug("Could not build notice message from template")
		return
	}

	// Send a notice to the Matrix rooms with the notice message.
	for _, room := range c.cfg.Notices.Rooms {
		_, err = c.c.SendNotice(room, b.String())

		// If there is was an error sending the notice to a specific room,
		// display the error without breaking from the loop in order to send the
		// notice to as much rooms possible.
		if err != nil {
			logEntry.Error(err)
		}
	}

	logEntry.Debug("Notice sent")

	return
}
