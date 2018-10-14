package types

// SCSData is a representation of a SCS, and is filled from both the data
// located in the SCS's PR and the strings located in the strings JSON file.
type SCSData struct {
	Number  int64
	Title   string
	Type    string
	State   string
	Message string
	URL     string
}
