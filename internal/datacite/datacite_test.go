package datacite

import (
	"encoding/json"
	"reflect"
	"testing"
)

type TestCase struct {
	Input []byte
	Want  DataciteRecord
}

var vanillaDataciteRecord DataciteRecord = DataciteRecord{
	Descriptions: []Description{
		{
			Description:     "Description",
			DescriptionType: "Abstract",
		},
	},
	Identifier: OutboundIdentifier{
		Value: "id",
	},
	Titles: []Title{
		{
			Title: "Title",
		},
	},
	Subjects: []Subject{
		{
			SchemeURI:     "schemeURI",
			Subject:       "subject",
			SubjectScheme: "subjectScheme",
			ValueURI:      "valueURI",
		},
	},
}

var happyPath TestCase = TestCase{
	Input: []byte(`{
				"id":"id",
				"type":"dois",
				"attributes":
					{
						"doi":"id",
						"identifiers":[
							{"identifier":"id","identifierType":"Publisher ID"}
						],
						"titles":[{"title":"Title"}],
						"subjects":[{"SchemeURI": "schemeURI", "Subject": "subject", "SubjectScheme": "subjectScheme", "ValueURI": "valueURI"}],
						"descriptions":[{"description":"Description","descriptionType":"Abstract"}]
					}
				}`),
	Want: vanillaDataciteRecord,
}

var multipleIdentifiers TestCase = TestCase{
	Input: []byte(`{
				"attributes":
					{
						"doi":"id",
						"identifiers":[
							{"identifier":["id","aac;AAC.00758-19v1"],"identifierType":"DOI"}
						],
						"titles":[{"title":"Title"}],
						"subjects":[{"SchemeURI": "schemeURI", "Subject": "subject", "SubjectScheme": "subjectScheme", "ValueURI": "valueURI"}],
						"descriptions":[{"description":"Description","descriptionType":"Abstract"}]
					}
				}`),
	Want: vanillaDataciteRecord,
}

func TestUnmarshalJSON(t *testing.T) {
	// Arrange
	table := []TestCase{
		happyPath,
		multipleIdentifiers,
	}
	// Act
	for _, tcase := range table {
		var got DataciteRecord
		err := json.Unmarshal(tcase.Input, &got)
		if err != nil {
			t.Errorf("Cannot unmarshal %v: %s", string(tcase.Input), err.Error())
			continue
		}
		// Assert
		if !reflect.DeepEqual(got, tcase.Want) {
			t.Errorf("%v is not equal to %v", got, tcase.Want)
		}
	}
}
