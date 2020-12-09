package httpx

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseForm(t *testing.T) {
	var v struct {
		Key      string `form:"key"`
		Location string `form:"location"`
		Radius   int64  `form:"radius" v:"min:1#半径不能为空"`
		Offset   int64  `form:"offset"`
		Page     int64  `form:"page"`
		Phone    string `form:"phone" v:"phone#手机号格式不正确"`
	}

	r, e := http.NewRequest(http.MethodGet, "http://localhost:8888/place/around?phone=18611914900&key=6e10597c6b5f745d2ff915a4a721edfb&location=116.473168,39.993015&radius=3000&extensions=base&output=json&offset=20&page=1", nil)
	if e = Parse(r, &v); e != nil {
		fmt.Println(e)
	}
	fmt.Println("key", v.Key)
	fmt.Println("location", v.Location)
	fmt.Println("radius", v.Radius)
	fmt.Println("offset", v.Offset)
	fmt.Println("page", v.Page)
	fmt.Println("phone", v.Phone)
}

func TestParseHeader(t *testing.T) {
	m := ParseHeader("key=value;")
	assert.EqualValues(t, map[string]string{
		"key": "value",
	}, m)
}

func TestParseJsonBody(t *testing.T) {
	var v struct {
		Name string `json:"name" v:"required"`
		Age  int    `json:"age"`
	}
	body := `{"name": "小王", age": 18}`

	//var v struct {
	//	Id string `json:"id" v:"required"`
	//}
	//body := `{"id": "1"}`

	r := httptest.NewRequest(http.MethodPost, "http://localhost:3333/", strings.NewReader(body))
	r.Header.Set(ContentType, ApplicationJson)

	if e := Parse(r, &v); e != nil {
		fmt.Println(e)
	}
	fmt.Println("id", v.Name)

	//assert.Nil(t, Parse(r, &v))
	//assert.Equal(t, "kevin", v.Name)
	//assert.Equal(t, 18, v.Age)
}
