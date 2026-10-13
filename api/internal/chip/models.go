package chip

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const MicroprocessorSourceWikipedia = "wikipedia"

// Details holds every other key of the document: they are stored and served at the top level, next to the fields below
type Microprocessor struct {
	Id           string              `json:"id" bson:"_id"`
	Name         string              `json:"name" bson:"name"`
	Type         string              `json:"type" bson:"type" example:"CPU"`
	Release      *primitive.DateTime `json:"release,omitempty" bson:"release,omitempty" swaggertype:"string" format:"date-time"`
	GateSize     float64             `json:"gateSize,omitempty" bson:"gate_size,omitempty"`
	TDP          float64             `json:"tdp,omitempty" bson:"tdp,omitempty"`
	DieSize      float64             `json:"dieSize,omitempty" bson:"die_size,omitempty"`
	Transistors  float64             `json:"transistors,omitempty" bson:"transistors,omitempty"`
	Density      float64             `json:"density,omitempty" bson:"density,omitempty"`
	Frequency    float64             `json:"frequency,omitempty" bson:"freq,omitempty"`
	Vendor       string              `json:"vendor,omitempty" bson:"vendor,omitempty"`
	Manufacturer string              `json:"manufacturer,omitempty" bson:"manufacturer,omitempty"`
	Source       string              `json:"source,omitempty" bson:"source,omitempty"`
	SourceURL    string              `json:"sourceUrl,omitempty" bson:"source_url,omitempty"`
	Details      map[string]any      `json:"-" bson:",inline"`
}

var microprocessorFields = map[string]bool{
	"_id": true, "name": true, "type": true, "release": true, "gate_size": true, "tdp": true, "die_size": true,
	"transistors": true, "density": true, "freq": true, "vendor": true, "manufacturer": true, "source": true, "source_url": true,
}

func (chip *Microprocessor) SetDetail(key string, value any) {
	if microprocessorFields[key] {
		return
	}
	if chip.Details == nil {
		chip.Details = map[string]any{}
	}
	chip.Details[key] = value
}

func (chip Microprocessor) MarshalJSON() ([]byte, error) {
	type fields Microprocessor
	encoded, err := json.Marshal(fields(chip))
	if err != nil {
		return nil, err
	}
	flat := map[string]any{}
	if err := json.Unmarshal(encoded, &flat); err != nil {
		return nil, err
	}
	for key, value := range chip.Details {
		if document, nested := value.(primitive.D); nested {
			value = document.Map()
		}
		if _, taken := flat[key]; !taken {
			flat[key] = value
		}
	}
	return json.Marshal(flat)
}

type MicroprocessorFilter struct {
	Type   string `form:"type,omitempty"`
	Vendor string `form:"vendor,omitempty"`
}
