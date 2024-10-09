package datacite

import (
	"encoding/json"
	"os"
	"regexp"

	"github.com/google/uuid"
)

var startsAndEndsWithBrackets = regexp.MustCompile(`^\[.*\]$`)

type DataciteRecord struct {
	Descriptions []Description      `json:"descriptions"`
	Identifier   OutboundIdentifier `json:"identifier"`
	Subjects     []Subject          `json:"subjects"`
	Titles       []Title            `json:"titles"`
}

type DatacitePayload struct {
	Documents []DataciteRecord `json:"documents"`
}

type Attributes struct {
	Identifiers  []InboundIdentifier `json:"identifiers"`
	Titles       []Title             `json:"titles"`
	Descriptions []Description       `json:"descriptions"`
	Subjects     []Subject           `json:"subjects"`
}

type DataCiteInboundPayload struct {
	Attributes   Attributes    `json:"attributes"`
	Descriptions []Description `json:"descriptions"`
	ID           string        `json:"id"`
	Identifiers  []Identifier  `json:"identifiers"`
	Subjects     []Subject     `json:"subjects"`
	Titles       []Title       `json:"titles"`
}

type OutboundIdentifier struct {
	Value string `json:"value"`
}

type InboundIdentifier struct {
	Identifier json.RawMessage `json:"identifier"`
	Type       string          `json:"identifierType"`
}

type Identifier struct {
	Identifier string `json:"identifier"`
	Type       string `json:"identifierType"`
}

type Title struct {
	Lang      string `json:"lang,omitempty"`
	Title     string `json:"title"`
	TitleType string `json:"titleType,omitempty"`
}

type Description struct {
	Description     string `json:"description,omitempty"`
	DescriptionType string `json:"descriptionType"`
	Lang            string `json:"lang,omitempty"`
}

type Subject struct {
	SchemeURI     string `json:"schemeUri,omitempty"`
	Subject       string `json:"subject"`
	SubjectScheme string `json:"subjectScheme,omitempty"`
	ValueURI      string `json:"valueUri,omitempty"`
}

func (d *DatacitePayload) Add(dataciteRecord DataciteRecord) {
	d.Documents = append(d.Documents, dataciteRecord)
}

func (d *DatacitePayload) Flush(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	json, err := json.Marshal(d)
	if err != nil {
		return err
	}
	_, err = f.Write(json)
	if err != nil {
		return err
	}
	d.Documents = d.Documents[:0]
	return nil
}

func (d *DataciteRecord) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}
	var inboundPayload DataCiteInboundPayload

	if err := json.Unmarshal(data, &inboundPayload); err != nil {
		return err
	}
	*d = DataciteRecord{
		Descriptions: inboundPayload.Descriptions,
		Identifier: OutboundIdentifier{
			Value: inboundPayload.ID,
		},
		Subjects: inboundPayload.Subjects,
		Titles:   inboundPayload.Titles,
	}
	if len(d.Descriptions) == 0 && len(inboundPayload.Attributes.Descriptions) > 0 {
		d.Descriptions = inboundPayload.Attributes.Descriptions
	}
	if len(d.Identifier.Value) == 0 {
		d.Identifier.Value = uuid.New().String()
		for _, id := range inboundPayload.Attributes.Identifiers {
			if id.Type != "DOI" || len(id.Identifier) == 0 {
				continue
			}
			if startsAndEndsWithBrackets.Match(id.Identifier) {
				identifiers := []string{}
				_ = json.Unmarshal(id.Identifier, &identifiers)
				d.Identifier.Value = identifiers[0]
			} else {
				var identifier string
				_ = json.Unmarshal(id.Identifier, &identifier)
				d.Identifier.Value = identifier
			}
		}
	}
	if len(d.Subjects) == 0 && len(inboundPayload.Attributes.Subjects) > 0 {
		d.Subjects = inboundPayload.Attributes.Subjects
	}
	if len(d.Titles) == 0 && len(inboundPayload.Attributes.Titles) > 0 {
		d.Titles = inboundPayload.Attributes.Titles
	}

	return nil
}

func (d *DataciteRecord) IsUseable() bool {
	for _, subject := range d.Subjects {
		if len(subject.SchemeURI) > 0 || len(subject.SubjectScheme) > 0 {
			return true
		}
	}
	return false
}
