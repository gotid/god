package queue

import (
	"errors"
	"github.com/gotid/god/lib/logx"
	"github.com/gotid/god/lib/mathx"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

var (
	proba    = mathx.NewProba()
	failProb = 0.01
)

func init() {
	logx.Disable()
}

func TestGenerateName(t *testing.T) {
	pushers := []Pusher{
		&mockedPusher{name: "first"},
		&mockedPusher{name: "second"},
		&mockedPusher{name: "third"},
	}

	assert.Equal(t, "first,second,third", generateName(pushers))
}

func TestGenerateNameNil(t *testing.T) {
	var pushers []Pusher
	assert.Equal(t, "", generateName(pushers))
}

func calcMean(vs []int) float64 {
	if len(vs) == 0 {
		return 0
	}

	var result float64
	for _, v := range vs {
		result += float64(v)
	}

	return result / float64(len(vs))
}

func calcVariance(mean float64, vs []int) float64 {
	if len(vs) == 0 {
		return 0
	}

	var result float64
	for _, v := range vs {
		result += math.Pow(float64(v)-mean, 2)
	}

	return result / float64(len(vs))
}

type mockedPusher struct {
	name  string
	count int
}

func (p *mockedPusher) Name() string {
	return p.name
}

func (p *mockedPusher) Push(message string) error {
	if proba.TrueOnProba(failProb) {
		return errors.New("dummy")
	}

	p.count++
	return nil
}
