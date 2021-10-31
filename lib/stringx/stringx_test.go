package stringx

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"git.zc0901.com/go/god/tools/god/mysql/builder"
	"github.com/stretchr/testify/assert"
)

func TestString_Split(t *testing.T) {
	dsn := "root:asdfasdf@tcp(192.168.0.166:3306)/dhome_project?parseTime=true"
	path := strings.Split(dsn, "?")[0]
	parts := strings.Split(path, "/")
	database := strings.TrimSpace(parts[len(parts)-1])
	if !strings.Contains(path, "/") || database == "" {
		t.Fatal("数据库连接字符串：未提供数据库名称")
	}
	fmt.Println(database)
}

func TestRemove(t *testing.T) {
	type Dict struct {
		Id             int64     `db:"id"`               // 字典表 | 公共库
		ParentId       int64     `db:"parent_id"`        // 字典分类ID
		Name           string    `db:"name"`             // 字典名称
		Type           string    `db:"type"`             // 字典类型
		CreateTime     time.Time `db:"create_time"`      // 创建时间
		CreateUserId   int64     `db:"create_user_id"`   // 创建人id
		CreateUserName string    `db:"create_user_name"` // 创建人姓名
		UpdateTime     time.Time `db:"update_time"`      // 更新时间
		UpdateUserId   int64     `db:"update_user_id"`   // 更新人id
		UpdateUserName string    `db:"update_user_name"` // 更新者姓名
		DeleteFlag     int64     `db:"delete_flag"`      // 删除标记: 0删除|1未删除
	}

	dictFieldList := builder.FieldList(&Dict{})
	dictFields := strings.Join(dictFieldList, ", ")
	dictFieldsAutoSet := strings.Join(Remove(dictFieldList, "id", "created_at", "updated_at", "create_time", "update_time"), ", ")
	fmt.Println(dictFields)
	fmt.Println(dictFieldsAutoSet)
}

func TestFilter(t *testing.T) {
	cases := []struct {
		input   string
		ignores []rune
		expect  string
	}{
		{``, nil, ``},
		{`abcd`, nil, `abcd`},
		{`ab,cd,ef`, []rune{','}, `abcdef`},
		{`ab, cd,ef`, []rune{',', ' '}, `abcdef`},
		{`ab, cd, ef`, []rune{',', ' '}, `abcdef`},
		{`ab, cd, ef `, []rune{',', ' '}, `abcdef`},
	}

	for _, each := range cases {
		t.Run(each.input, func(t *testing.T) {
			actual := Filter(each.input, func(r rune) bool {
				for _, ignore := range each.ignores {
					if ignore == r {
						return true
					}
				}
				return false
			})
			assert.Equal(t, each.expect, actual)
		})
	}
}

func TestA(t *testing.T) {
	fields := strings.FieldsFunc("ab, cd, ef ", func(r rune) bool {
		return r == ',' || r == ' '
	})
	fmt.Println(fields)
}
