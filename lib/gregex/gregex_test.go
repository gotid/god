package gregex_test

import (
	"testing"

	"github.com/gotid/god/lib/gregex"
	"github.com/stretchr/testify/assert"
)

var PatternErr = `([\d+`

func Test_Quote(t *testing.T) {
	s1 := `[foo]` //`\[foo\]`
	assert.Equal(t, gregex.Quote(s1), `\[foo\]`)
}

func Test_Validate(t *testing.T) {
	s1 := `(.+):(\d+)`
	assert.Equal(t, gregex.Validate(s1), nil)
	s1 = `((.+):(\d+)`
	assert.Equal(t, gregex.Validate(s1) == nil, false)
}

func Test_IsMatch(t *testing.T) {
	pattern := `(.+):(\d+)`
	s1 := []byte(`sfs:2323`)
	assert.Equal(t, gregex.IsMatch(pattern, s1), true)
	s1 = []byte(`sfs2323`)
	assert.Equal(t, gregex.IsMatch(pattern, s1), false)
	s1 = []byte(`sfs:`)
	assert.Equal(t, gregex.IsMatch(pattern, s1), false)
	// error pattern
	assert.Equal(t, gregex.IsMatch(PatternErr, s1), false)
}

func Test_IsMatchString(t *testing.T) {
	pattern := `(.+):(\d+)`
	s1 := `sfs:2323`
	assert.Equal(t, gregex.IsMatchString(pattern, s1), true)
	s1 = `sfs2323`
	assert.Equal(t, gregex.IsMatchString(pattern, s1), false)
	s1 = `sfs:`
	assert.Equal(t, gregex.IsMatchString(pattern, s1), false)
	// error pattern
	assert.Equal(t, gregex.IsMatchString(PatternErr, s1), false)
}

func Test_Match(t *testing.T) {
	re := "a(a+b+)b"
	wantSubs := "aaabb"
	s := "acbb" + wantSubs + "dd"
	subs, err := gregex.Match(re, []byte(s))
	assert.Equal(t, err, nil)
	if string(subs[0]) != wantSubs {
		t.Fatalf("regex:%s,Match(%q)[0] = %q; want %q", re, s, subs[0], wantSubs)
	}
	if string(subs[1]) != "aab" {
		t.Fatalf("Match(%q)[1] = %q; want %q", s, subs[1], "aab")
	}
	// error pattern
	_, err = gregex.Match(PatternErr, []byte(s))
	assert.NotEqual(t, err, nil)
}

func Test_MatchString(t *testing.T) {
	re := "a(a+b+)b"
	wantSubs := "aaabb"
	s := "acbb" + wantSubs + "dd"
	subs, err := gregex.MatchString(re, s)
	assert.Equal(t, err, nil)
	if string(subs[0]) != wantSubs {
		t.Fatalf("regex:%s,Match(%q)[0] = %q; want %q", re, s, subs[0], wantSubs)
	}
	if string(subs[1]) != "aab" {
		t.Fatalf("Match(%q)[1] = %q; want %q", s, subs[1], "aab")
	}
	// error pattern
	_, err = gregex.MatchString(PatternErr, s)
	assert.NotEqual(t, err, nil)
}
