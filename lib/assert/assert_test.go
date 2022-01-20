package assert

import (
	"fmt"
	"testing"

	"github.com/gotid/god/internal/empty"
)

type RelationshipType string

func (r RelationshipType) String() string {
	return string(r)
}

const All RelationshipType = ""

func TestIsNotEmpty(t *testing.T) {
	fmt.Println(fmt.Sprintf("%v", ""), empty.IsEmpty(""))
	fmt.Println(fmt.Sprintf("%v", All), empty.IsEmpty(All))
	fmt.Println(fmt.Sprintf("%v", 0), empty.IsEmpty(0))
	fmt.Println(fmt.Sprintf("%v", false), empty.IsEmpty(false))
}
