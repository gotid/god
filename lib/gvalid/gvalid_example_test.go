package gvalid_test

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/gotid/god/lib/g"
	"github.com/gotid/god/lib/gconv"
	"github.com/gotid/god/lib/gvalid"
)

func TestCheckMap(t *testing.T) {
	params := g.Map{
		"passport":  "",
		"password":  "123456",
		"password2": "1234567",
	}

	rules := []string{
		"passport@required|length:6,16#账号不能为空|账号长度应当在:min到:max之间",
		"password@required|length:6,16|same:password2#密码不能为空|密码长度应当在:min到:max之间|两次密码输入不相等",
		"password2@required|length:6,16#",
	}

	if e := gvalid.CheckMap(params, rules); e != nil {
		fmt.Println(e.Map())
		fmt.Println(e.FirstItem())
		fmt.Println(e.FirstString())
		fmt.Println(e.FirstRule())
	}
}

func TestCheckStruct(t *testing.T) {
	type Params struct {
		Page      int `v:"required|min:1	# 页码必填"`
		Size      int `v:"required|between:1,100 # 每页条数必填"`
		ProjectId int `v:"between:1,10000 # 项目编号必须在 :min 和 :max 之间"`
	}

	params := &Params{
		Page: 1,
		Size: 10,
	}

	err := gvalid.CheckStruct(params, nil)
	fmt.Println(err)
}

func TestRegisterRule(t *testing.T) {
	// 自定义校验规则（唯一名称）
	rule := "unique-name"
	_ = gvalid.RegisterRule(rule, func(rule string, value interface{}, message string, params map[string]interface{}) error {
		name := gconv.String(value)
		return errors.New("用户名称 " + name + " 已存在")
	})

	// 校验测试
	type User struct {
		Id   int
		Name string `v:"required|unique-name#请输入用户名称|用户名称已被占用"`
		Pass string `v:"required|length:6,18"`
	}
	u := &User{
		Id:   1,
		Name: "mark",
		Pass: "123455",
	}
	err := gvalid.CheckStruct(u, nil)
	fmt.Println(err.Error())
}

func TestRegisterRule_OverwriteRequired(t *testing.T) {
	// 重写 required 验证规则
	rule := "required-strict"
	_ = gvalid.RegisterRule(rule, func(rule string, value interface{}, message string, params map[string]interface{}) error {
		reflectValue := reflect.ValueOf(value)
		if reflectValue.Kind() == reflect.Ptr {
			reflectValue = reflectValue.Elem()
		}
		isEmpty := false
		switch reflectValue.Kind() {
		case reflect.Bool:
			isEmpty = !reflectValue.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			isEmpty = reflectValue.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			isEmpty = reflectValue.Uint() == 0
		case reflect.Float32, reflect.Float64:
			isEmpty = math.Float64bits(reflectValue.Float()) == 0
		case reflect.Complex64, reflect.Complex128:
			c := reflectValue.Complex()
			isEmpty = math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0
		case reflect.String, reflect.Map, reflect.Array, reflect.Slice:
			isEmpty = reflectValue.Len() == 0
		}
		if isEmpty {
			return errors.New(message)
		}
		return nil
	})

	fmt.Println(1, gvalid.Check("", "required-strict", "该项必填"))
	fmt.Println(2, gvalid.Check([]string{}, "required-strict", "该项必填"))
	fmt.Println(3, gvalid.Check(g.SliceStr{}, "required-strict", "该项必填"))
	fmt.Println(4, gvalid.Check(g.MapStrInt{}, "required-strict", "该项必填"))
	gvalid.DeleteRule(rule)
	fmt.Println()
	fmt.Println(1, gvalid.Check("", "required-strict", "该项必填"))
	fmt.Println(2, gvalid.Check([]string{}, "required-strict", "该项必填"))
	fmt.Println(3, gvalid.Check(g.SliceStr{}, "required-strict", "该项必填"))
	fmt.Println(4, gvalid.Check(g.MapStrInt{}, "required-strict", "该项必填"))
}
