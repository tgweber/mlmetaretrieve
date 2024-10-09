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

func TestUnmarshalJSON(t *testing.T) {
	// Arrange
	table := []TestCase{
		{
			Input: []byte(`{
				"id":"10.1128/aac.00758-19",
				"type":"dois",
				"attributes":
					{
						"doi":"10.1128/aac.00758-19",
						"identifiers":[
							{"identifier":["AAC.00758-19","aac;AAC.00758-19v1"],"identifierType":"Publisher ID"}
						],
						"titles":[{"title":"Title"}],
						"subjects":[],
						"descriptions":[{"description":"Description","descriptionType":"Abstract"}]
					}
				}`),
			Want: DataciteRecord{
				Descriptions: []Description{
					{
						Description:     "Description",
						DescriptionType: "Abstract",
					},
				},
				Identifier: OutboundIdentifier{
					Value: "10.1128/aac.00758-19",
				},
				Titles: []Title{
					{
						Title: "Title",
					},
				},
				Subjects: []Subject{},
			},
		},
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
