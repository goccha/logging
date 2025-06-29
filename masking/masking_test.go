package masking

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"testing"

	"github.com/gorilla/schema"
	"github.com/stretchr/testify/assert"
)

func TestProcessor_JsonMask(t *testing.T) {
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

	data, err := New("password", "name", "age", "phone", "address", "introducers").Json(ctx, body)
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

func TestJson(t *testing.T) {
	str := "{\"username\":\"test_user\",\"password\":\"qwerty\"}"
	ctx := context.Background()
	data, err := New("password").Json(ctx, str)
	assert.NoError(t, err)
	body := map[string]interface{}{}
	err = json.Unmarshal(data, &body)
	assert.NoError(t, err)
	assert.Equal(t, MaskValue, body["password"])

	str = ""
	data, err = New("password").Json(ctx, str)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(data))
}

func TestProcessor_FormMask(t *testing.T) {
	type UserPass struct {
		Username string `schema:"username" json:"username"`
		Password string `schema:"password" json:"password"`
		Range    []int  `schema:"range" json:"range"`
	}
	req := &UserPass{
		Username: "test_user",
		Password: "qwerty",
		Range:    []int{10, 20},
	}
	encoder := schema.NewEncoder()
	body := url.Values{}
	err := encoder.Encode(req, body)
	assert.NoError(t, err)

	ctx := context.Background()
	data, err := New("password").Form(ctx, base64.URLEncoding.EncodeToString([]byte(body.Encode())))
	assert.NoError(t, err)

	form, err := url.ParseQuery(string(data))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(form["password"]))
	assert.Equal(t, MaskValue, form["password"][0])
}

func TestForm(t *testing.T) {
	str := "username=test_user&password=qwerty"
	ctx := context.Background()
	data, err := New("password").Form(ctx, base64.URLEncoding.EncodeToString([]byte(str)))
	assert.NoError(t, err)
	form, err := url.ParseQuery(string(data))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(form["password"]))
	assert.Equal(t, MaskValue, form["password"][0])

	str = ""
	data, err = New("password").Form(ctx, str)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(data))
}
