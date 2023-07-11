package hashicorpdocs
//custom-template-add
// BaseDoc contains common document metadata fields used by Hermes.
type BaseTemplate struct {
	// ObjectID is the Google Drive file ID for the document.
	TemplateId string `json:"templateID"`
	Description string `json:"description"`
	LongName string `json:"longName"`
	DocId string `json:"docId"`
}