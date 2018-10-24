package matrix

import (
	"fmt"
	"strings"
	"text/template"

	"config"
	"types"

	"github.com/matrix-org/gomatrix"
)

// Cli is a representation of a Matrix client, containing both the gomatrix
// client and the configuration.
type Cli struct {
	c   *gomatrix.Client
	cfg *config.Config
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
	return
}

// SendNoticeWithTypeAndState generates a notice message from the SCS data and
// then sends the said message as a notice to the configured Matrix rooms.
// Returns an error if the message could not be generated or if the notice could
// not be sent to the Matrix rooms.
// Returns and do nothing if there's no message string available for the SCS's
// SCSP state.
func (c *Cli) SendNoticeWithTypeAndState(data types.SCSData) (err error) {
	// Check if there's a message string available for the given SCS type and
	// SCSP state.
	var ok bool
	data.Message, ok = c.cfg.Notices.Strings[data.Type][data.State]
	if !ok {
		// If a message string could not be found for the given SCS type and
		// SCSP state, check if there's a type-independant message string for
		// this SCSP state.
		data.Message, ok = c.cfg.Notices.Strings["global"][data.State]

		// If no string could be found, return and do nothing.
		if !ok {
			return
		}
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
	data types.SCSData, unsplitLabels []string,
) (err error) {
	var ok bool
	var match string
	for _, l := range unsplitLabels {
		// If we have another match when there's already a message loaded in,
		// we don't know what message to use. In this case, don't do anything.
		if match, ok = c.cfg.Notices.Strings["global"][l]; ok && len(data.Message) > 0 {
			return
		}

		data.Message = match
	}

	// No message could be found for any of the labels.
	if len(data.Message) == 0 {
		return
	}

	return c.sendNotice(data)
}

// sendNotice uses the given data to generate the full notice message for this
// submission update from the configured template, and send it to the Matrix
// rooms.
// Returns with an error it there was an issue generating the notice message
// from the configured template, or sending it out as a notice to the Matrix
// room.
func (c *Cli) sendNotice(data types.SCSData) (err error) {
	// Load the template defined in the configuration file. The "message" name
	// used here is not important.
	tmpl, err := template.New("message").Parse(c.cfg.Notices.Pattern)
	if err != nil {
		return
	}

	// Generate the notice message from the configured template and the SCS's
	// data.
	var b strings.Builder
	if err = tmpl.Execute(&b, data); err != nil {
		return
	}

	// Send a notice to the Matrix rooms with the notice message.
	for _, room := range c.cfg.Notices.Rooms {
		_, err = c.c.SendNotice(room, b.String())

		// If there is was an error sending the notice to a specific room,
		// display the error without breaking from the loop in order to send the
		// notice to as much rooms possible.
		if err != nil {
			fmt.Println(err)
		}
	}

	return
}
