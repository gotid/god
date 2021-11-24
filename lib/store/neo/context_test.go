package neo

import (
	"fmt"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"

	"git.zc0901.com/go/god/lib/g"

	"github.com/stretchr/testify/assert"
)

var ctx *Context

func init() {
	ctx = MustContext(cfg)
}

func TestContext_Driver(t *testing.T) {
	assert.Equal(t, "bolt", ctx.Driver().Target().Scheme)
	assert.Equal(t, "localhost:7687", ctx.Driver().Target().Host)
}

func TestConfig_RunCypher(t *testing.T) {
	cypher := `match (input:Person)-[r1:FRIEND_OF]->(through:Person)-[r2:LIVES_IN]->(city:City)<-[r3:LIVES_IN]-(reco:Person)
				where id(input)=0 and id(input) <> id(reco)
				return distinct id(reco) as reco
				limit $limit`

	t.Run("未制定回调函数", func(t *testing.T) {
		err := ctx.RunCypher(cypher, nil, g.Map{"limit": cfg.Limit})
		assert.Equal(t, ErrNoScanner, err)
	})

	t.Run("有参数却未指定参数值", func(t *testing.T) {
		err := ctx.RunCypher(cypher, callback)
		assert.NotNil(t, err)
	})

	t.Run("指定回调和参数值应正常运行", func(t *testing.T) {
		err := ctx.RunCypher(cypher, callback, g.Map{"limit": 2})
		assert.Nil(t, err)
	})
}

func BenchmarkAutoTx(b *testing.B) {
	for i := 0; i < b.N; i++ {
		query()
	}
}

func BenchmarkManualTx(b *testing.B) {
	tx, err := ctx.BeginTx()
	defer tx.Close()
	assert.Nil(b, err)
	for i := 0; i < b.N; i++ {
		queryWithTx(tx)
	}
}

func callback(result neo4j.Result) error {
	var ids []int64
	for result.Next() {
		record := result.Record()
		if v, ok := record.Get("reco"); ok {
			ids = append(ids, v.(int64))
		}
	}
	fmt.Println(ids)
	return nil
}

func query() {
	cypher := "match (p:Person) return id(p) as reco limit 10"
	ctx.RunCypher(cypher, callback)
}

func queryWithTx(tx neo4j.Transaction) {
	cypher := "match (p:Person) return id(p) as reco limit 10"
	ctx.RunCypherWithTx(tx, cypher, callback)
}
