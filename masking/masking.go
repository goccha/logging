package masking

import (
	"context"
	"encoding/json"
	"strings"
)

var MaskValue = "*****"
var MaskNumber = 0

type Processor struct {
	keys  []string
	Value string
}

func New(keys ...string) *Processor {
	return &Processor{
		keys:  keys,
		Value: MaskValue,
	}
}

func (b *Processor) Add(key string, keys ...string) *Processor {
	b.keys = append(b.keys, key)
	for _, k := range keys {
		b.keys = append(b.keys, k)
	}
	return b
}
func (b *Processor) GetMaskValue() string {
	return b.Value
}
func (b *Processor) SetMaskValue(value string) {
	b.Value = value
}

func (b *Processor) Run(ctx context.Context, data []byte) ([]byte, error) {
	body := make(map[string]interface{})
	if err := json.Unmarshal(data, &body); err != nil {
		return data, err
	}
	body = b.masking(ctx, body)
	return json.Marshal(body)
}

func (b *Processor) masking(ctx context.Context, body map[string]interface{}) map[string]interface{} {
	for k, v := range body {
		if contains(b.keys, k) {
			switch val := v.(type) {
			case string:
				if val != "" {
					body[k] = b.Value
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
							val[i] = b.Value
						}
					case float64:
						if iv != 0 {
							val[i] = float64(MaskNumber)
						}
					}

				}
			}
		} else {
			if m, ok := v.(map[string]interface{}); ok {
				body[k] = b.masking(ctx, m)
			} else if a, ok := v.([]interface{}); ok {
				for i, av := range a {
					switch iv := av.(type) {
					case map[string]interface{}:
						a[i] = b.masking(ctx, iv)
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
