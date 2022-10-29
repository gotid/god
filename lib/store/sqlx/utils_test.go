package sqlx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		args   []interface{}
		expect string
		hasErr bool
	}{
		{
			name:   "mysql normal",
			query:  "select name, age from users where bool=? and phone=?",
			args:   []interface{}{true, "133"},
			expect: "select name, age from users where bool=1 and phone='133'",
		},
		{
			name:   "mysql normal",
			query:  "select name, age from users where bool=? and phone=?",
			args:   []interface{}{false, "133"},
			expect: "select name, age from users where bool=0 and phone='133'",
		},
		{
			name:   "pg normal",
			query:  "select name, age from users where bool=$1 and phone=$2",
			args:   []interface{}{true, "133"},
			expect: "select name, age from users where bool=1 and phone='133'",
		},
		{
			name:   "pg normal reverse",
			query:  "select name, age from users where bool=$2 and phone=$1",
			args:   []interface{}{"133", false},
			expect: "select name, age from users where bool=0 and phone='133'",
		},
		{
			name:   "pg error not number",
			query:  "select name, age from users where bool=$a and phone=$1",
			args:   []interface{}{"133", false},
			hasErr: true,
		},
		{
			name:   "pg error more args",
			query:  "select name, age from users where bool=$2 and phone=$1 and nickname=$3",
			args:   []interface{}{"133", false},
			hasErr: true,
		},
		{
			name:   "oracle normal",
			query:  "select name, age from users where bool=:1 and phone=:2",
			args:   []interface{}{true, "133"},
			expect: "select name, age from users where bool=1 and phone='133'",
		},
		{
			name:   "oracle normal reverse",
			query:  "select name, age from users where bool=:2 and phone=:1",
			args:   []interface{}{"133", false},
			expect: "select name, age from users where bool=0 and phone='133'",
		},
		{
			name:   "oracle error not number",
			query:  "select name, age from users where bool=:a and phone=:1",
			args:   []interface{}{"133", false},
			hasErr: true,
		},
		{
			name:   "oracle error more args",
			query:  "select name, age from users where bool=:2 and phone=:1 and nickname=:3",
			args:   []interface{}{"133", false},
			hasErr: true,
		},
		{
			name:   "select with date",
			query:  "select * from user where date='2006-01-02 15:04:05' and name=:1",
			args:   []interface{}{"foo"},
			expect: "select * from user where date='2006-01-02 15:04:05' and name='foo'",
		},
		{
			name:   "select with date and escape",
			query:  `select * from user where date=' 2006-01-02 15:04:05 \'' and name=:1`,
			args:   []interface{}{"foo"},
			expect: `select * from user where date=' 2006-01-02 15:04:05 \'' and name='foo'`,
		},
		{
			name:   "select with date and bad arg",
			query:  `select * from user where date='2006-01-02 15:04:05 \'' and name=:a`,
			args:   []interface{}{"foo"},
			hasErr: true,
		},
		{
			name:   "select with date and escape error",
			query:  `select * from user where date='2006-01-02 15:04:05 \`,
			args:   []interface{}{"foo"},
			hasErr: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			actual, err := format(test.query, test.args...)
			if test.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expect, actual)
			}
		})
	}
}
