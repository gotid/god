package neo

import (
	"fmt"
	"strings"

	"git.zc0901.com/go/god/lib/g"
)

// Relation 表示一个带方向的关系边。
type Relation struct {
	Type      RelationType
	Direction Direction
	Params    g.Map
}

// NewRelation 返回一个新的关系边。
func NewRelation(t RelationType, d Direction, m ...g.Map) Relation {
	r := Relation{
		Type:      t,
		Direction: d,
	}
	if len(m) > 0 {
		r.Params = m[0]
	}
	return r
}

// Edge 返回关系特征字符串。
//
// alias 用于指定关系别名。
//
// 返回结果形如： -[:VIEW]-> 或 <-[r:VIEW]-
func (r *Relation) Edge(alias ...string) string {
	return r.edge(alias, "")
}

// EdgeWithParams 返回关系边特征字符串（带有参数）。
//
// alias 用于指定关系别名。
//
// 返回结果形如： -[:VIEW {time:123]-> 或 <-[r:VIEW {time:123]-
func (r *Relation) EdgeWithParams(alias ...string) string {
	return r.edge(alias, " "+r.parseParams())
}

// OnSet 返回关系设置字符串。
//
// alias 用于指定关系别名。
//
// 返回结果形如： -[:VIEW]->, ON CREATE SET ..., ON MATCH SET ...
func (r *Relation) OnSet(alias string) string {
	onCreateSet := fmt.Sprintf("ON CREATE SET %s=%s", alias, r.parseParams())
	onMergeSet := fmt.Sprintf("ON MATCH SET %s", r.parseParams2(alias))
	return onCreateSet + "\n" + onMergeSet
}

func (r *Relation) edge(alias []string, params string) string {
	var ali string
	if len(alias) == 1 {
		ali = alias[0]
	}
	switch r.Direction {
	case Outgoing:
		return r.edgeString("", ali, r.Type, params, ">")
	case Incoming:
		return r.edgeString("<", ali, r.Type, params, "")
	case Both:
		return r.edgeString("", ali, r.Type, params, "")
	default:
		return ""
	}
}

// 返回关系的边特征字符串。
func (r *Relation) edgeString(left, alias string, relType RelationType, params, right string) string {
	typ := relType
	if typ != "" {
		typ = ":" + typ
	}

	return fmt.Sprintf("%s-[%s%s%s]-%s", left, alias, typ, params, right)
}

// 解析用于neo4j中key不包括双引号的字典。
func (r *Relation) parseParams() string {
	if r.Params == nil {
		return ""
	}

	b := strings.Builder{}
	b.WriteString("{")
	index := 0
	for k, v := range r.Params {
		index++
		b.WriteString(k)
		b.WriteString(":")
		switch v.(type) {
		case string:
			b.WriteString(fmt.Sprintf(`"%s"`, v))
		default:
			b.WriteString(fmt.Sprintf("%v", v))
		}
		if index != len(r.Params) {
			b.WriteRune(',')
		}
	}
	b.WriteString("}")
	return b.String()
}

// 解析用于neo4j中key不包括双引号的字典。
func (r *Relation) parseParams2(alias string) string {
	if r.Params == nil {
		return ""
	}

	b := strings.Builder{}
	index := 0
	for k, v := range r.Params {
		index++
		b.WriteString(fmt.Sprintf("%s.%s=", alias, k))
		switch v.(type) {
		case string:
			b.WriteString(fmt.Sprintf(`"%s"`, v))
		default:
			b.WriteString(fmt.Sprintf("%v", v))
		}
		if index != len(r.Params) {
			b.WriteRune(',')
		}
	}
	return b.String()
}
