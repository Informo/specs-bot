package matrix

import (
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
// then sends the said message as a notice to the configured Matrix room.
// Returns an error if the message could not be generated or if the notice could
// not be sent to the Matrix room.
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

func (c *Cli) SendNoticeWithUnsplitLabels(
	data types.SCSData, unsplitLabels []string,
) (err error) {
	var ok bool
	for _, l := range unsplitLabels {
		if len(data.Message) > 0 && _, ok = c.cfg.Notices.Strings["global"][l]; ok {
			return
		}

		data.Message, ok = c.cfg.Notices.Strings["global"][l]

		// If no string could be found, return and do nothing.
		if !ok {
			return
		}
	}

	return c.sendNotice(data)
}

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

	// Send a notice to the Matrix room with the notice message.
	_, err = c.c.SendNotice(c.cfg.Notices.Room, b.String())

	return
}
