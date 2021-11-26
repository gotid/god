package neo

import "fmt"

// Relationship 表示一个带方向的关系边。
type Relationship struct {
	Type      RelationshipType
	Direction Direction
}

// NewRelationship 返回一个新的关系边。
func NewRelationship(t RelationshipType, d Direction) *Relationship {
	return &Relationship{
		Type:      t,
		Direction: d,
	}
}

// Edge 返回由关系类型和方向组成的边特征字符串。
//
// alias 用于指定别名
//
// 返回结果形如： -[:FOLLOW]-> 或 <-[f:FOLLOW]-
func (r *Relationship) Edge(alias ...string) string {
	var ali string
	if len(alias) > 0 {
		ali = alias[0]
	}
	switch r.Direction {
	case Outgoing:
		return r.edge("", ali, r.Type, ">")
	case Incoming:
		return r.edge("<", ali, r.Type, "")
	case Both:
		return r.edge("", ali, r.Type, "")
	}
	return ""
}

func (r *Relationship) edge(left, alias string, relType RelationshipType, right string) string {
	typ := relType
	if typ != "" {
		typ = ":" + typ
	}
	return fmt.Sprintf("%s-[%s%s]-%s", left, alias, typ, right)
}
