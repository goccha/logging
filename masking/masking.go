package masking

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/gorilla/schema"
)

var MaskValue = "*****"
var MaskNumber = 0

type Processor struct {
	keys []string
}

func New(keys ...string) *Processor {
	return &Processor{
		keys: keys,
	}
}

func (b *Processor) Add(key string, keys ...string) *Processor {
	b.keys = append(b.keys, key)
	b.keys = append(b.keys, keys...)
	return b
}

var empty []byte

// Json converts a JSON string or struct to JSON bytes.
func (b *Processor) Json(ctx context.Context, v any) (data []byte, err error) {
	var str string
	if _, ok := v.([]byte); ok {
		str = string(v.([]byte))
	} else if _, ok := v.(string); ok {
		str = v.(string)
	} else {
		if data, err = json.Marshal(v); err != nil {
			return nil, err
		}
		str = string(data)
	}
	if len(str) == 0 {
		return empty, nil
	}
	if str[0] == '[' { // JSON array
		body := make([]map[string]interface{}, 0)
		if err = json.Unmarshal([]byte(str), &body); err != nil {
			return data, err
		}
		if data, err = json.Marshal(maskingArray(ctx, body, b.keys)); err != nil {
			return data, err
		}
		return data, nil
	} else {
		body := make(map[string]interface{})
		if err = json.Unmarshal([]byte(str), &body); err != nil {
			return data, err
		}
		if data, err = json.Marshal(masking(ctx, body, b.keys)); err != nil {
			return data, err
		}
		return data, nil
	}
}

var encoder = schema.NewEncoder()

// Form converts a form-encoded string to URL-encoded bytes, masking sensitive values.
func (b *Processor) Form(ctx context.Context, v any) (data []byte, err error) {
	var form url.Values
	var str string
	switch val := v.(type) {
	case []byte:
		str = string(val)
	case string:
		str = val
	default:
		if err = encoder.Encode(v, form); err != nil {
			return nil, err
		}
	}
	if len(form) == 0 {
		if len(str) == 0 {
			return empty, nil
		}
		if bin, err := base64.URLEncoding.DecodeString(str); err != nil {
			return nil, err
		} else {
			str = string(bin)
		}
		if form, err = url.ParseQuery(str); err != nil {
			return
		}
	}
	return []byte(maskingForm(ctx, form, b.keys).Encode()), nil
}

// maskingForm processes a form-encoded body and masks values based on the keys provided.
func maskingForm(ctx context.Context, body url.Values, keys []string) url.Values {
	for k, v := range body {
		for _, s := range keys {
			if _, ok := body[s]; ok { // Exact match
				for i, val := range v {
					if val != "" {
						body[k][i] = MaskValue
					}
				}
			} else if strings.EqualFold(k, s) { // Case-insensitive match
				for i, val := range v {
					if val != "" {
						body[k][i] = MaskValue
					}
				}
			}
		}
	}
	return body
}

// maskingArray processes an array of maps and masks values based on the keys provided.
func maskingArray(ctx context.Context, body []map[string]interface{}, keys []string) []map[string]interface{} {
	for i, item := range body {
		body[i] = masking(ctx, item, keys)
	}
	return body
}

// masking processes the map and masks values based on the keys provided.
func masking(ctx context.Context, body map[string]interface{}, keys []string) map[string]interface{} {
	for k, v := range body {
		if contains(keys, k) {
			switch val := v.(type) {
			case string:
				if val != "" {
					body[k] = MaskValue
				}
			case float64:
				if val != 0 {
					body[k] = MaskNumber
				}
			case []interface{}:
				for i, iv := range val {
					switch vv := iv.(type) {
					case string:
						if vv != "" {
							val[i] = MaskValue
						}
					case float64:
						if iv != 0 {
							val[i] = MaskNumber
						}
					}
				}
			}
		} else {
			switch val := v.(type) {
			case map[string]interface{}:
				body[k] = masking(ctx, val, keys)
			case []interface{}:
				for i, av := range val {
					switch iv := av.(type) {
					case map[string]interface{}:
						val[i] = masking(ctx, iv, keys)
					}
				}
			}
		}
	}
	return body
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if strings.EqualFold(s, v) {
			return true
		}
	}
	return false
}
