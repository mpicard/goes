package graph

import (
	"errors"
	"io"
	"time"

	"github.com/vektah/gqlgen/graphql"
)

// if the type referenced in types.json is a function that returns a marshaller we can use it to encode and decode
// onto any existing go type.
func MarshalTimestamp(t time.Time) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		io.WriteString(w, "\""+t.UTC().Format(time.RFC3339)+"\"")
	})
}

// Unmarshal{Typename} is only required if the scalar appears as an input. The raw values have already been decoded
// from json into int/float64/bool/nil/map[string]interface/[]interface
func UnmarshalTimestamp(v interface{}) (time.Time, error) {
	if tmpStr, ok := v.(string); ok {
		return time.Parse(time.RFC3339, tmpStr)
	}
	return time.Time{}, errors.New("time should be a RFC3339 timestamp")
}
