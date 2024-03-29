package mathx

import (
	"github.com/gotid/god/lib/stringx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMaxInt(t *testing.T) {
	cases := []struct {
		a      int
		b      int
		expect int
	}{
		{
			a:      0,
			b:      1,
			expect: 1,
		},
		{
			a:      0,
			b:      -1,
			expect: 0,
		},
		{
			a:      1,
			b:      1,
			expect: 1,
		},
	}

	for _, each := range cases {
		each := each
		t.Run(stringx.Rand(), func(t *testing.T) {
			actual := MaxInt(each.a, each.b)
			assert.Equal(t, each.expect, actual)
		})
	}
}

func TestMax(t *testing.T) {
	cases := []struct {
		a      float64
		b      float64
		expect float64
	}{
		{
			a:      0,
			b:      1,
			expect: 1,
		},
		{
			a:      0,
			b:      -1,
			expect: 0,
		},
		{
			a:      1,
			b:      1,
			expect: 1,
		},
	}

	for _, each := range cases {
		each := each
		t.Run(stringx.Rand(), func(t *testing.T) {
			actual := Max([]float64{each.a, each.b})
			assert.Equal(t, each.expect, actual)
		})
	}
}

func TestMinInt(t *testing.T) {
	cases := []struct {
		a      int
		b      int
		expect int
	}{
		{
			a:      0,
			b:      1,
			expect: 0,
		},
		{
			a:      0,
			b:      -1,
			expect: -1,
		},
		{
			a:      1,
			b:      1,
			expect: 1,
		},
	}

	for _, each := range cases {
		t.Run(stringx.Rand(), func(t *testing.T) {
			actual := MinInt(each.a, each.b)
			assert.Equal(t, each.expect, actual)
		})
	}
}

func TestMin(t *testing.T) {
	cases := []struct {
		a      float64
		b      float64
		expect float64
	}{
		{
			a:      0,
			b:      1,
			expect: 0,
		},
		{
			a:      0,
			b:      -1,
			expect: -1,
		},
		{
			a:      1,
			b:      1,
			expect: 1,
		},
	}

	for _, each := range cases {
		t.Run(stringx.Rand(), func(t *testing.T) {
			actual := Min([]float64{each.a, each.b})
			assert.Equal(t, each.expect, actual)
		})
	}
}
