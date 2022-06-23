package masking

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcessor_Mask(t *testing.T) {
	type Profile struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Gender  string `json:"gender"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}
	type Body struct {
		UserId      string `json:"userId"`
		Password    string `json:"password"`
		Profile     `json:"profile"`
		Families    []Profile `json:"profiles"`
		Introducers []string  `json:"introducers"`
	}

	body := &Body{
		UserId:   "test_user",
		Password: "password123",
		Profile: Profile{
			Name:    "山田 太郎",
			Age:     30,
			Gender:  "男",
			Phone:   "00-0123-4567",
			Address: "東京都",
		},
		Families: []Profile{
			{
				Name:    "山田 花子",
				Age:     30,
				Gender:  "女",
				Phone:   "00-0123-4568",
				Address: "東京都",
			},
			{
				Name:    "山田 一",
				Age:     3,
				Gender:  "男",
				Phone:   "",
				Address: "東京都",
			},
		},
		Introducers: []string{"田中 太郎", "鈴木 一郎"},
	}
	ctx := context.Background()

	data, err := json.Marshal(body)
	assert.NoError(t, err)

	data, err = New("password", "name", "age", "phone", "address", "introducers").Run(ctx, data)
	assert.NoError(t, err)

	err = json.Unmarshal(data, body)
	assert.NoError(t, err)

	assert.Equal(t, MaskValue, body.Password)
	assert.Equal(t, MaskValue, body.Name)
	assert.Equal(t, MaskNumber, body.Age)
	assert.Equal(t, MaskValue, body.Phone)
	assert.Equal(t, MaskValue, body.Address)
	assert.Equal(t, "男", body.Gender)
	assert.Equal(t, "test_user", body.UserId)

	for i, v := range body.Families {
		if i == 0 {
			assert.Equal(t, MaskValue, v.Name)
			assert.Equal(t, MaskNumber, v.Age)
			assert.Equal(t, MaskValue, v.Phone)
			assert.Equal(t, MaskValue, v.Address)
		} else {
			assert.Equal(t, MaskValue, v.Name)
			assert.Equal(t, MaskNumber, v.Age)
			assert.Equal(t, "", v.Phone)
			assert.Equal(t, MaskValue, v.Address)
		}
	}
	for _, v := range body.Introducers {
		assert.Equal(t, MaskValue, v)
	}
}
