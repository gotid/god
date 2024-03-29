package fx

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRetry(t *testing.T) {
	assert.NotNil(t, DoWithRetry(func() error {
		print("重试\n")
		return errors.New("any")
	}))

	var times int
	assert.Nil(t, DoWithRetry(func() error {
		times++
		if times == defaultRetryTimes {
			return nil
		}
		return errors.New("any")
	}))

	times = 0
	assert.NotNil(t, DoWithRetry(func() error {
		times++
		if times == defaultRetryTimes+1 {
			return nil
		}
		return errors.New("any")
	}))

	total := 2 * defaultRetryTimes
	times = 0
	assert.Nil(t, DoWithRetry(func() error {
		times++
		if times == total {
			return nil
		}
		return errors.New("any")
	}, WithRetry(total)))
}
