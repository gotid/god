package queue

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiPusher(t *testing.T) {
	const pusherCount = 100
	var pushers []Pusher
	var mockedPushers []*mockedPusher
	for i := 0; i < pusherCount; i++ {
		p := &mockedPusher{
			name: "pusher:" + strconv.Itoa(i),
		}
		pushers = append(pushers, p)
		mockedPushers = append(mockedPushers, p)
	}

	pusher := NewMultiPusher(pushers)
	assert.True(t, len(pusher.Name()) > 0)

	for i := 0; i < 1000; i++ {
		_ = pusher.Push("item")
	}

	var counts []int
	for _, p := range mockedPushers {
		counts = append(counts, p.count)
	}
	mean := calcMean(counts)
	variance := calcVariance(mean, counts)
	assert.True(t, math.Abs(mean-1000*(1-failProb)) < 10)
	assert.True(t, variance < 100, fmt.Sprintf("差异太大 - %.2f", variance))
}
