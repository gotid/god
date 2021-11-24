package neo

import (
	"fmt"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"

	"github.com/stretchr/testify/assert"
)

func TestNewNeo(t *testing.T) {
	neo := NewNeo(target, username, password, "")

	t.Run("单值——简单类型测试", func(t *testing.T) {
		var tomId int64
		err := neo.Read2(&tomId, `MATCH (tom:Person {name: "Tom Hanks"}) RETURN id(tom)`)
		assert.Nil(t, err)
		assert.Equal(t, int64(71), tomId)
	})

	t.Run("单值——结构体测试", func(t *testing.T) {
		var tom struct {
			Id   int64  `neo:"id"`
			Name string `neo:"name"`
		}
		err := neo.Read2(&tom, `MATCH (tom:Person {name: "Tom Hanks"})-[:ACTED_IN]->() RETURN id(tom) as id, tom.name as name`)
		assert.Nil(t, err)
		assert.Equal(t, int64(71), tom.Id)
		assert.Equal(t, "Tom Hanks", tom.Name)
	})

	t.Run("单值——结构体测试2", func(t *testing.T) {
		var result struct {
			Tom neo4j.Node `neo:"tom"`
		}
		err := neo.Read2(&result, `MATCH (tom:Person {name: "Tom Hanks"})-[:ACTED_IN]->() RETURN tom`)
		assert.Nil(t, err)
		assert.Equal(t, int64(71), result.Tom.Id)
		assert.Equal(t, "Tom Hanks", result.Tom.Props["name"])
	})

	t.Run("多值——简单类型测试", func(t *testing.T) {
		var names []string
		err := neo.Read2(&names, `MATCH (cloudAtlas:Movie {title: "Cloud Atlas"})<-[:DIRECTED]-(directors) RETURN directors.name`)
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
		err := neo.Read2(&tomMovies, `MATCH (tom:Person {name: "Tom Hanks"})-[:ACTED_IN]->(tomHanksMovie) RETURN tomHanksMovie as movie`)
		assert.Nil(t, err)
		for i, movie := range tomMovies {
			fmt.Println(i, movie.Movie)
		}
	})
}
