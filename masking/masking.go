package masking

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
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

func (b *Processor) Run(ctx context.Context, f Unmarshal) ([]byte, error) {
	data, err := f()
	if err != nil {
		return nil, err
	}
	body := make(map[string]interface{})
	if err = json.Unmarshal(data, &body); err != nil {
		return data, err
	}
	body = b.masking(ctx, body)
	return json.Marshal(body)
}

type Unmarshal func() ([]byte, error)

func Raw(data []byte) Unmarshal {
	return func() ([]byte, error) {
		return data, nil
	}
}

func Json(v interface{}) Unmarshal {
	return func() (data []byte, err error) {
		if str, ok := v.(string); ok {
			return []byte(str), nil
		}
		return json.Marshal(v)
	}
}

func Form(v interface{}) Unmarshal {
	return func() (data []byte, err error) {
		switch val := v.(type) {
		case string:
			var form url.Values
			if form, err = url.ParseQuery(val); err != nil {
				return
			}
			return json.Marshal(form)
		default:
			return json.Marshal(val)
		}
	}
}

func (b *Processor) masking(ctx context.Context, body map[string]interface{}) map[string]interface{} {
	for k, v := range body {
		if contains(b.keys, k) {
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
				body[k] = b.masking(ctx, val)
			case []interface{}:
				for i, av := range val {
					switch iv := av.(type) {
					case map[string]interface{}:
						val[i] = b.masking(ctx, iv)
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
