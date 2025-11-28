package json5

import (
	"encoding/json"
	"io"
)

// Marshal returns the JSON encoding of v.
// It delegates to encoding/json.Marshal.
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalIndent is like Marshal but applies Indent to format the output.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// Unmarshal parses the JSON5-encoded data and stores the result
// in the value pointed to by v.
func Unmarshal(data []byte, v interface{}) error {
	transcoded, err := Transcode(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(transcoded, v)
}

// Decoder reads and decodes JSON5 values from an input stream.
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the next JSON5-encoded value from its
// input and stores it in the value pointed to by v.
// Note: Currently this reads the entire remaining stream to transcode.
func (dec *Decoder) Decode(v interface{}) error {
	data, err := io.ReadAll(dec.r)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return io.EOF
	}
	return Unmarshal(data, v)
}

// Encoder writes JSON values to an output stream.
type Encoder struct {
	enc *json.Encoder
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{enc: json.NewEncoder(w)}
}

// Encode writes the JSON encoding of v to the stream.
func (enc *Encoder) Encode(v interface{}) error {
	return enc.enc.Encode(v)
}

// SetIndent instructs the encoder to format each subsequent encoded
// value as if marshaled by MarshalIndent.
func (enc *Encoder) SetIndent(prefix, indent string) {
	enc.enc.SetIndent(prefix, indent)
}

// SetEscapeHTML specifies whether problematic HTML characters
// should be escaped inside JSON quoted strings.
func (enc *Encoder) SetEscapeHTML(on bool) {
	enc.enc.SetEscapeHTML(on)
}
