package neo

import (
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

func BenchmarkOriginalRunCypher(b *testing.B) {
	cypher := `MATCH (tom:Person {name: "Tom Hanks"})-[:ACTED_IN]->() RETURN tom`

	b.Run("原生自动事务测试", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := ctx.RunCypher(cypher, callback)
			assert.Nil(b, err)
		}
	})

	b.Run("原生手动事务测试", func(b *testing.B) {
		tx, err := ctx.BeginTx()
		defer tx.Close()
		assert.Nil(b, err)
		for i := 0; i < b.N; i++ {
			err := ctx.RunCypherWithTx(tx, cypher, callback)
			assert.Nil(b, err)
		}
	})
}

func callback(result neo4j.Result) error {
	return nil
}
