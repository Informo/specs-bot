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

// SendNotice generates a notice message from the SCS data and then sends the
// said message as a notice to the configured Matrix rooms.
// Returns an error if the message could not be generated or if the notice could
// not be sent to the Matrix rooms.
// Returns and do nothing if there's no message string available for the SCS's
// SCSP state.
func (c *Cli) SendNotice(data types.SCSData) (err error) {
	// Load the template defined in the configuration file. The "message" name
	// used here is not important.
	tmpl, err := template.New("message").Parse(c.cfg.Notices.Pattern)
	if err != nil {
		return
	}

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

	// Generate the notice message from the configured template and the SCS's
	// data.
	var b strings.Builder
	if err = tmpl.Execute(&b, data); err != nil {
		return
	}

	// Send a notice to the Matrix rooms with the notice message.
	for _, room := range c.cfg.Notices.Rooms {
		_, err = c.c.SendNotice(room, b.String())

		// If there is an error for one specific room, display the error and continue
		// looping
		if err != nil {
			fmt.Println(err)
		}
	}

	return
}
