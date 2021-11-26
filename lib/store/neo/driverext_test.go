package neo

import (
	"fmt"
	"testing"

	"git.zc0901.com/go/god/lib/fx"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func TestFxParallel(t *testing.T) {
	type ln struct {
		Labels string
		Nodes  []*neo4j.Node
	}

	nodes := []*neo4j.Node{
		{Id: 1, Labels: []string{"User"}},
		{Id: 2, Labels: []string{"User"}},
		{Id: 3, Labels: []string{"Project"}},
		{Id: 4, Labels: []string{"Project"}},
		{Id: 5, Labels: []string{"Project"}},
	}
	nodeMap := groupNodes(nodes)
	fmt.Println(nodeMap)
	fx.From(func(source chan<- interface{}) {
		for ls, ns := range nodeMap {
			source <- []interface{}{ls, ns}
		}
	}).Parallel(func(item interface{}) {
		vs := item.([]interface{})
		ns := vs[1].([]*neo4j.Node)
		for _, n := range ns {
			fmt.Println(vs[0], n.Id)
		}
	})
}
