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
		UserId   string `json:"userId"`
		Password string `json:"password"`
		Profile  `json:"profile"`
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
	}
	ctx := context.Background()

	data, err := json.Marshal(body)
	assert.NoError(t, err)

	data, err = New("password", "name", "age", "phone", "address").Run(ctx, data)
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

}
