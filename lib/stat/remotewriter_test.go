package stat

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"testing"
)

func TestRemoteWriter(t *testing.T) {
	defer gock.Off()

	gock.New("https://foo.com").Reply(200).BodyString("foo")
	writer := NewRemoteWriter("https://foo.com")
	err := writer.Write(&StatReport{
		Name: "bar",
	})
	assert.Nil(t, err)
}

func TestRemoteWriterFail(t *testing.T) {
	defer gock.Off()

	gock.New("https://foo.com").Reply(503).BodyString("foo")
	writer := NewRemoteWriter("https://foo.com")
	err := writer.Write(&StatReport{
		Name: "bar",
	})
	assert.NotNil(t, err)
}
