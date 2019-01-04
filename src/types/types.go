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

// CopyWithMsg returns a new instance of SCSData with the given string as its
// Message field.
func (d *SCSData) CopyWithMsg(msg string) *SCSData {
	newData := new(SCSData)

	newData.Number = d.Number
	newData.Title = d.Title
	newData.Type = d.Type
	newData.State = d.State
	newData.URL = d.URL

	newData.Message = msg

	return newData
}
