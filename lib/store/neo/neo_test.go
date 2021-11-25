package neo

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

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

var neo = NewNeo(target, username, password, "")

func TestNewNeo(t *testing.T) {
	SetSlowThreshold(10 * time.Millisecond)

	t.Run("单值——简单类型测试", func(t *testing.T) {
		var tomId int64
		err := neo.Read(&tomId, `MATCH (tom:User {Nick: "苗雨露"}) RETURN id(tom)`)
		assert.Nil(t, err)
		assert.Equal(t, int64(5), tomId)
	})

	t.Run("单值——结构体测试", func(t *testing.T) {
		var tom struct {
			Id   int64  `neo:"id"`
			Name string `neo:"name"`
		}
		err := neo.Read(&tom, `MATCH (tom:Person {name: "Tom Hanks"})-[:ACTED_IN]->() RETURN id(tom) as id, tom.name as name`)
		assert.Nil(t, err)
		assert.Equal(t, int64(71), tom.Id)
		assert.Equal(t, "Tom Hanks", tom.Name)
	})

	t.Run("单值——结构体测试2", func(t *testing.T) {
		var result struct {
			Node neo4j.Node `neo:"n"`
		}
		err := neo.Read(&result, `MATCH (n:User) WHERE n.Id=318 RETURN n LIMIT 25`)
		assert.Nil(t, err)
		fmt.Println(result)
		// assert.Equal(t, int64(71), result.Tom.Id)
		// assert.Equal(t, "Tom Hanks", result.Tom.Props["name"])
	})

	t.Run("多值——简单类型测试", func(t *testing.T) {
		var names []string
		err := neo.Read(&names, `MATCH (cloudAtlas:Movie {title: "Cloud Atlas"})<-[:DIRECTED]-(directors) RETURN directors.name`)
		assert.Nil(t, err)
		assert.Len(t, names, 3)
		assert.EqualValues(t,
			[]string{"Tom Tykwer", "Lana Wachowski", "Lilly Wachowski"},
			names,
		)
	})

	t.Run("多值——结构体测试", func(t *testing.T) {
		type Movie struct {
			Movie neo4j.Node `neo:"movie"`
		}
		var tomMovies []Movie
		err := neo.Read(&tomMovies, `MATCH (tom:Person {name: "Tom Hanks"})-[:ACTED_IN]->(tomHanksMovie) RETURN tomHanksMovie as movie`)
		assert.Nil(t, err)
		for i, movie := range tomMovies {
			fmt.Println(i, movie.Movie)
		}
	})

	t.Run("最短路径查询-读数", func(t *testing.T) {
		cypher := `MATCH p=shortestPath((bacon:Person {name:"Kevin Bacon"})-[*]-(meg:Person {name:"Meg Ryan"})) RETURN p`
		var dest struct {
			Path neo4j.Path `neo:"p"`
		}
		err := neo.Read(&dest, cypher)
		assert.Nil(t, err)
		for _, node := range dest.Path.Nodes {
			fmt.Println("节点", node)
		}
		for _, r := range dest.Path.Relationships {
			fmt.Println("关系", r)
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
		err := neo.Run(cb, cypher)
		assert.Nil(t, err)
	})

	t.Run("创建唯一约束测试", func(t *testing.T) {
		err := neo.Run(nil, `create constraint unq_project_id if not exists on (n:Project) assert n.Id is unique`)
		assert.Nil(t, err)
	})

	t.Run("获取 schema", func(t *testing.T) {
		var dest interface{}
		err := neo.Read(&dest, "drop constraint constraint_1ea8c423")
		assert.NotNil(t, err)
	})

	t.Run("托管式事务测试", func(t *testing.T) {
		err := neo.Transact(func(tx neo4j.Transaction) error {
			var dest interface{}
			err := neo.Read(&dest, "drop constraint constraint_1ea8c423")
			if err != nil {
				return err
			}

			return nil
		})
		assert.NotNil(t, err)
	})
}

func TestDriver_SingleOtherNode(t *testing.T) {
	input := &neo4j.Node{Id: 1}
	rel := NewRelationship("All", Outgoing)
	otherNode, err := neo.SingleOtherNode(input, rel)
	assert.Nil(t, err)
	fmt.Println(otherNode)
}

func TestDriver_CreateNode(t *testing.T) {
	id := int64(6)
	err := neo.CreateNode(&neo4j.Node{
		Id:     id,
		Labels: []string{"User", "Project"},
		Props: map[string]interface{}{
			"id":       id,
			"nickname": "苗雨露",
		},
	})
	assert.Nil(t, err)
}

func BenchmarkMergeNode(b *testing.B) {
	logx.Disable()

	nodes := make([]*neo4j.Node, 0)
	labels := []string{"User", "Project"}

	for i := 0; i < b.N; i++ {
		n := rand.Intn(2)
		id := rand.Int63()
		label := labels[n]
		nodes = append(nodes, &neo4j.Node{
			Id:     id,
			Labels: []string{label},
			Props: map[string]interface{}{
				"id":   id,
				"name": stringx.RandId(),
			},
		})
	}

	fmt.Println("本批次数量 ", len(nodes))
	err := neo.MergeNode(nodes...)
	assert.Nil(b, err)
}

func BenchmarkRunCypherWithBreaker(b *testing.B) {
	logx.SetLevel(logx.ErrorLevel)

	var result struct {
		Tom neo4j.Node `neo:"tom"`
	}

	cypher := `MATCH (tom:Person {name: "Tom Hanks"})-[:ACTED_IN]->() RETURN tom`

	query := func() {
		err := neo.Read(&result, cypher)
		assert.Nil(b, err)
	}

	txQuery := func(tx neo4j.Transaction) {
		err := neo.TxRead(tx, &result, cypher)
		assert.Nil(b, err)
	}

	b.Run("断路器版自动事务测试", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			query()
		}
	})

	b.Run("托管的事务性操作测试", func(b *testing.B) {
		err := neo.Transact(func(tx neo4j.Transaction) error {
			for i := 0; i < b.N; i++ {
				txQuery(tx)
			}
			return nil
		})
		assert.Nil(b, err)
	})

	tx, err := neo.BeginTx()
	assert.Nil(b, err)
	defer tx.Close()
	b.Run("断路器版手动事务测试", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			txQuery(tx)
		}
	})
}
