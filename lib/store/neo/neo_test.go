package neo

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"git.zc0901.com/go/god/lib/g"

	"git.zc0901.com/go/god/lib/stringx"

	"git.zc0901.com/go/god/lib/logx"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"

	"github.com/stretchr/testify/assert"
)

const (
	target   = "bolt://localhost:7687"
	username = "neo4j"
	password = "asdfasdf"
)

var (
	neo = NewNeo(target, username, password, "")
	ctx = Context{}
)

func TestTx(t *testing.T) {
	err := neo.Transact(func(tx neo4j.Transaction) error {
		var id int64
		err := neo.Read(ctx, &id, `MATCH (u:User {id: 318}) RETURN u.id`)
		assert.Nil(t, err)
		return nil
	})
	assert.Nil(t, err)
}

func TestNoRecords(t *testing.T) {
	var result struct {
		Node neo4j.Node `neo:"n"`
	}

	cypher := `MATCH (n:Person {name: "Tom Hanks"})-[:ACTED_IN]->() RETURN n limit 1`
	err := neo.Read(ctx, &result, cypher)
	assert.NotNil(t, err)

	cypher = `MATCH (n:User {id: 318}) RETURN n limit 1`
	err = neo.Read(ctx, &result, cypher)
	assert.Nil(t, err)
	fmt.Println(result)
}

func TestNewNeo(t *testing.T) {
	SetSlowThreshold(10 * time.Millisecond)

	t.Run("单值——简单类型测试", func(t *testing.T) {
		var id int64
		err := neo.Read(ctx, &id, `MATCH (u:User {id: 318}) RETURN u.id`)
		assert.Nil(t, err)
		assert.Equal(t, int64(318), id)
	})

	t.Run("单值——结构体测试", func(t *testing.T) {
		var user struct {
			Id   int64  `neo:"id"`
			Name string `neo:"name"`
		}
		err := neo.Read(ctx, &user, `MATCH (u:User {id: 318}) RETURN u.id as id, u.nickname as name`)
		assert.Nil(t, err)
		assert.Equal(t, int64(318), user.Id)
		assert.Equal(t, "自在", user.Name)
	})

	t.Run("单值——结构体测试2", func(t *testing.T) {
		var user struct {
			Node neo4j.Node `neo:"n"`
		}
		err := neo.Read(ctx, &user, `MATCH (i)-[:VIEW]->(n) WHERE i.id=6 RETURN n limit 1`)
		assert.Nil(t, err)
		fmt.Println(user)
	})

	t.Run("多值——简单类型测试", func(t *testing.T) {
		var names []string
		cypher := `MATCH (cloudAtlas:Movie {title: "Cloud Atlas"})<-[:DIRECTED]-(directors) RETURN directors.name`
		err := neo.Read(ctx, &names, cypher)
		assert.Nil(t, err)
		fmt.Println(names)
		//assert.Len(t, names, 3)
		//assert.EqualValues(t,
		//	[]string{"Tom Tykwer", "Lana Wachowski", "Lilly Wachowski"},
		//	names,
		//)
	})

	t.Run("多值——结构体测试", func(t *testing.T) {
		type Movie struct {
			Movie neo4j.Node `neo:"movie"`
		}
		var tomMovies []Movie
		cypher := `MATCH (tom:Person {name: "Tom Hanks"})-[:ACTED_IN]->(tomHanksMovie) RETURN tomHanksMovie as movie`
		err := neo.Read(ctx, &tomMovies, cypher)
		assert.Nil(t, err)
		for i, movie := range tomMovies {
			fmt.Println(i, movie.Movie)
		}
	})

	t.Run("最短路径查询-读数", func(t *testing.T) {
		cypher := `MATCH p=shortestPath((bacon:Person {name:"Kevin Bacon"})-[*]-(meg:Person {name:"Meg Ryan"})) RETURN p`
		var paths []struct {
			Path neo4j.Path `neo:"p"`
		}
		for _, path := range paths {
			err := neo.Read(ctx, &path, cypher)
			assert.Nil(t, err)
			for _, node := range path.Path.Nodes {
				fmt.Println("节点", node)
			}
			for _, r := range path.Path.Relationships {
				fmt.Println("关系", r)
			}
		}
	})

	t.Run("最短路径查询 - 运行", func(t *testing.T) {
		cb := func(result neo4j.Result) error {
			for result.Next() {
				record := result.Record()
				p, ok := record.Get("p")
				fmt.Printf("%#v\n", record)
				if !ok {
					continue
				}
				fmt.Printf("%#v\n", p)
			}
			return nil
		}
		cypher := `MATCH p=shortestPath((bacon:Person {name:"Kevin Bacon"})-[*]-(meg:Person {name:"Meg Ryan"})) RETURN p`
		err := neo.Run(ctx, cb, cypher)
		assert.Nil(t, err)
	})

	t.Run("创建唯一约束测试", func(t *testing.T) {
		err := neo.Run(ctx, nil, `create constraint unq_project_id if not exists on (n:Project) assert n.Id is unique`)
		assert.Nil(t, err)
	})

	t.Run("获取 schema", func(t *testing.T) {
		var dest interface{}
		err := neo.Read(ctx, &dest, "drop constraint constraint_1ea8c423")
		assert.NotNil(t, err)
	})

	//t.Run("托管式事务测试", func(t *testing.T) {
	//	err := neo.Transact(ctx, func(tx neo4j.Transaction) error {
	//		var dest interface{}
	//		err := neo.Read(ctx, &dest, "drop constraint constraint_1ea8c423")
	//		if err != nil {
	//			return err
	//		}
	//
	//		return nil
	//	})
	//	assert.NotNil(t, err)
	//})
}

func TestRelationship_Edge(t *testing.T) {
	r := NewRelation(View, Both)
	assert.Equal(t, "-[:VIEW]-", r.Edge())

	r = NewRelation(View, Incoming)
	assert.Equal(t, "<-[r:VIEW]-", r.Edge("r"))

	r = NewRelation(View, Outgoing, g.Map{
		"name":    "zs",
		"age":     123,
		"enabled": true,
		"time":    time.Now().Unix(),
	})
	fmt.Println("仅关系边", r.Edge("r"))
	fmt.Println("带参关系边", r.EdgeWithParams("r"))

	onset := r.OnSet("r")
	fmt.Println(onset)
}

func TestDriver_SingleOtherNode(t *testing.T) {
	input := neo4j.Node{Id: 6}
	rel := NewRelation(Down, Outgoing)
	otherNode, err := neo.SingleOtherNode(ctx, input, rel)
	assert.Nil(t, err)
	fmt.Println(otherNode)

	rel = NewRelation(View, Outgoing)
	otherNode, err = neo.SingleOtherNode(ctx, input, rel)
	assert.NotNil(t, err)
	fmt.Println(otherNode)
}

func TestDriver_GetDegree(t *testing.T) {
	t.Run("全部双向关系数量", func(t *testing.T) {
		input := neo4j.Node{Id: 6}
		rel := NewRelation(All, Both)
		degree, err := neo.GetDegree(ctx, input, rel)
		assert.Nil(t, err)
		assert.Equal(t, int64(4), degree)
	})

	t.Run("双向浏览数量", func(t *testing.T) {
		input := neo4j.Node{Id: 6}
		rel := NewRelation(View, Both)
		degree, err := neo.GetDegree(ctx, input, rel)
		assert.Nil(t, err)
		assert.Equal(t, int64(3), degree)
	})

	t.Run("浏览数量", func(t *testing.T) {
		input := neo4j.Node{Id: 6}
		rel := NewRelation(View, Outgoing)
		degree, err := neo.GetDegree(ctx, input, rel)
		assert.Nil(t, err)
		assert.Equal(t, int64(2), degree)
	})

	t.Run("被下载数量", func(t *testing.T) {
		input := neo4j.Node{Id: 319}
		rel := NewRelation(Down, Incoming)
		degree, err := neo.GetDegree(ctx, input, rel)
		assert.Nil(t, err)
		assert.Equal(t, int64(1), degree)
	})
}

func TestDriver_CreateNode(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	id := rand.Int63()
	err := neo.CreateNode(ctx, neo4j.Node{
		Id:     id,
		Labels: []string{"User", "Project"},
		Props: map[string]interface{}{
			"id":       id,
			"nickname": "自自在在",
		},
	})
	assert.Nil(t, err)
}

func TestDriver_MergeNode(t *testing.T) {
	t.Run("批量替换节点", func(t *testing.T) {
		err := neo.ReplaceNodes(ctx, neo4j.Node{
			Id:     331,
			Labels: []string{"User", "Project"},
			Props: map[string]interface{}{
				"name": "朱邵",
			},
		})
		assert.Nil(t, err)
	})

	t.Run("单个合成节点", func(t *testing.T) {
		err := neo.MergeNode(ctx, neo4j.Node{
			Id:     332,
			Labels: []string{"User"},
			Props: map[string]interface{}{
				"star": 5,
				"time": time.Now(),
			},
		})
		assert.Nil(t, err)
	})
}

func TestDriver_CreateRelation(t *testing.T) {
	n1 := neo4j.Node{Id: 1, Labels: []string{"User"}}
	r := NewRelation("FOLLOW", Outgoing, g.Map{"time": time.Now().Unix()})
	n2 := neo4j.Node{Id: 2, Labels: []string{"User"}}

	err := neo.CreateRelation(ctx, n1, r, n2)
	assert.Nil(t, err)
}

func TestDriver_DeleteRelation(t *testing.T) {
	n1 := neo4j.Node{Id: 1, Labels: []string{"User"}}
	r := NewRelation("FOLLOW", Outgoing, g.Map{"time": time.Now().Unix()})
	n2 := neo4j.Node{Id: 2, Labels: []string{"User"}}

	err := neo.DeleteRelation(ctx, n1, r, n2)
	assert.Nil(t, err)
}

func BenchmarkMergeNode(b *testing.B) {
	logx.Disable()

	nodes := make([]neo4j.Node, 0)
	labels := []string{"User", "Project"}

	for i := 0; i < b.N; i++ {
		n := rand.Intn(2)
		id := rand.Int63()
		label := labels[n]
		nodes = append(nodes, neo4j.Node{
			Id:     id,
			Labels: []string{label},
			Props: map[string]interface{}{
				"id":   id,
				"name": stringx.RandId(),
			},
		})
	}

	fmt.Println("本批次数量 ", len(nodes))
	err := neo.ReplaceNodes(ctx, nodes...)
	assert.Nil(b, err)
}

func BenchmarkRunCypherWithBreaker(b *testing.B) {
	logx.SetLevel(logx.ErrorLevel)

	var result struct {
		Tom neo4j.Node `neo:"tom"`
	}

	cypher := `MATCH (tom:Person {name: "Tom Hanks"})-[:ACTED_IN]->() RETURN tom limit 1`

	query := func() {
		err := neo.Read(ctx, &result, cypher)
		assert.Nil(b, err)
	}

	//txQuery := func(tx neo4j.Transaction) {
	//	err := neo.Read(ctx, &result, cypher)
	//	assert.Nil(b, err)
	//}

	b.Run("断路器版自动事务测试", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			query()
		}
	})

	b.Run("托管的事务性操作测试", func(b *testing.B) {
		err := neo.Transact(func(tx neo4j.Transaction) error {
			ctx.Tx = tx
			for i := 0; i < b.N; i++ {
				err := neo.Read(ctx, &result, cypher)
				assert.Nil(b, err)
			}

			return nil
		})
		assert.Nil(b, err)
	})

	tx, err := neo.BeginTx()
	assert.Nil(b, err)
	defer tx.Close()
	ctx.Tx = tx
	b.Run("断路器版手动事务测试", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// txQuery(tx)
			query()
		}
	})
}
