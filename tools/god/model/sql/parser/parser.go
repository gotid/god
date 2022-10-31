package parser

import (
	"fmt"
	"github.com/gotid/ddl-parser/parser"
	"github.com/gotid/god/lib/collection"
	"github.com/gotid/god/tools/god/model/sql/converter"
	"github.com/gotid/god/tools/god/model/sql/model"
	"github.com/gotid/god/tools/god/model/sql/util"
	"github.com/gotid/god/tools/god/util/console"
	"github.com/gotid/god/tools/god/util/stringx"
	"path/filepath"
	"sort"
	"strings"
)

const timeImport = "time.Time"

type (
	// Table 表示一张数据表。
	Table struct {
		Name        stringx.String
		Db          stringx.String
		PrimaryKey  Primary
		UniqueIndex map[string][]*Field
		Fields      []*Field
		ContainsPQ  bool
	}

	// Primary 表示一个自增主键字段。
	Primary struct {
		Field
		AutoIncrement bool
	}

	// Field 表示一个表字段。
	Field struct {
		OriginalName    string
		Name            stringx.String
		DataType        string
		Comment         string
		SeqInIndex      int
		OrdinalPosition int
		ContainsPQ      bool
	}

	// KeyType 是 int 的别名。
	KeyType int
)

// ContainsTime 判断是否包含 golang 类型 time.Time。
func (t *Table) ContainsTime() bool {
	for _, field := range t.Fields {
		if field.DataType == timeImport {
			return true
		}
	}

	return false
}

// ConvertDataType 将 mysql 数据类型转为 golang 数据类型。
func ConvertDataType(table *model.Table, strict bool) (*Table, error) {
	isPrimaryDefaultNull := table.PrimaryKey.ColumnDefault == nil && table.PrimaryKey.IsNullAble == "YES"
	isPrimaryUnsigned := strings.Contains(table.PrimaryKey.DbColumn.ColumnType, "unsigned")
	primaryDataType, containPQ, err := converter.ConvertStringDataType(table.PrimaryKey.DataType, isPrimaryDefaultNull, isPrimaryUnsigned, strict)
	if err != nil {
		return nil, err
	}

	seqInIndex := 0
	if table.PrimaryKey.Index != nil {
		seqInIndex = table.PrimaryKey.Index.SeqInIndex
	}
	reply := Table{
		Name: stringx.From(table.Table),
		Db:   stringx.From(table.Db),
		PrimaryKey: Primary{
			Field: Field{
				OriginalName:    "",
				Name:            stringx.From(table.PrimaryKey.Name),
				DataType:        primaryDataType,
				Comment:         table.PrimaryKey.Comment,
				SeqInIndex:      seqInIndex,
				OrdinalPosition: table.PrimaryKey.OrdinalPosition,
			},
			AutoIncrement: strings.Contains(table.PrimaryKey.Extra, "auto_increment"),
		},
		UniqueIndex: map[string][]*Field{},
		ContainsPQ:  containPQ,
	}

	fieldM, err := getTableFields(table, strict)
	if err != nil {
		return nil, err
	}

	for _, field := range fieldM {
		if field.ContainsPQ {
			reply.ContainsPQ = true
		}
		reply.Fields = append(reply.Fields, field)
	}

	sort.Slice(reply.Fields, func(i, j int) bool {
		return reply.Fields[i].OrdinalPosition < reply.Fields[j].OrdinalPosition
	})

	uniqueIndexSet := collection.NewSet()
	log := console.NewColorConsole()
	for indexName, columns := range table.UniqueIndex {
		sort.Slice(columns, func(i, j int) bool {
			if columns[i].Index != nil {
				return columns[i].Index.SeqInIndex < columns[j].Index.SeqInIndex
			}

			return false
		})

		if len(columns) == 1 {
			one := columns[0]
			if one.Name == table.PrimaryKey.Name {
				log.Warning("[ConvertDataType]：表 %q，主键唯一索引不可重复：%q", table.Table, one.Name)
				continue
			}
		}

		var fields []*Field
		var uniqueJoin []string
		for _, col := range columns {
			fields = append(fields, fieldM[col.Name])
			uniqueJoin = append(uniqueJoin, col.Name)
		}

		uniqueKey := strings.Join(uniqueJoin, ",")
		if uniqueIndexSet.Contains(uniqueKey) {
			log.Warning("[ConvertDataType]：表 %q，唯一索引重复：%q", table.Table, uniqueKey)
			continue
		}

		uniqueIndexSet.AddStr(uniqueKey)
		reply.UniqueIndex[indexName] = fields
	}

	return &reply, nil
}

// Parse 解析 ddl 脚本至 golang 结构体。
func Parse(filename, database string, strict bool) ([]*Table, error) {
	p := parser.NewParser()
	tables, err := p.From(filename)
	if err != nil {
		return nil, err
	}

	originalNames := parseOriginalName(tables)
	indexNameGen := func(column ...string) string {
		return strings.Join(column, "_")
	}

	prefix := filepath.Base(filename)
	var list []*Table
	for indexTable, table := range tables {
		var (
			primaryColumn    string
			primaryColumnSet = collection.NewSet()
			uniqueKeyMap     = make(map[string][]string)
			normalKeyMap     = make(map[string][]string)
			columns          = table.Columns
		)

		for _, column := range columns {
			if column.Constraint != nil {
				if column.Constraint.Primary {
					primaryColumnSet.AddStr(column.Name)
				}

				if column.Constraint.Unique {
					indexName := indexNameGen(column.Name, "unique")
					uniqueKeyMap[indexName] = []string{column.Name}
				}

				if column.Constraint.Key {
					indexName := indexNameGen(column.Name, "idx")
					uniqueKeyMap[indexName] = []string{column.Name}
				}
			}
		}

		for _, constraint := range table.Constraints {
			if len(constraint.ColumnPrimaryKey) > 1 {
				return nil, fmt.Errorf("%s：只能有一个主键，不支持联合主键", prefix)
			}

			if len(constraint.ColumnPrimaryKey) == 1 {
				primaryColumn = constraint.ColumnPrimaryKey[0]
				primaryColumnSet.AddStr(constraint.ColumnPrimaryKey[0])
			}

			if len(constraint.ColumnUniqueKey) > 0 {
				uniqueKeys := append([]string(nil), constraint.ColumnUniqueKey...)
				uniqueKeys = append(uniqueKeys, "unique")
				indexName := indexNameGen(uniqueKeys...)
				uniqueKeyMap[indexName] = constraint.ColumnUniqueKey
			}
		}

		if primaryColumnSet.Count() > 1 {
			return nil, fmt.Errorf("%s：只能有一个主键，不支持联合主键", prefix)
		}

		primaryKey, fieldM, err := convertColumns(columns, primaryColumn, strict)
		if err != nil {
			return nil, err
		}

		var fields []*Field
		// 排序
		for indexColumn, column := range columns {
			field, ok := fieldM[column.Name]
			if ok {
				field.OriginalName = originalNames[indexTable][indexColumn]
				fields = append(fields, field)
			}
		}

		var (
			uniqueIndex = make(map[string][]*Field)
			normalIndex = make(map[string][]*Field)
		)
		for indexName, each := range uniqueKeyMap {
			for _, columnName := range each {
				uniqueIndex[indexName] = append(uniqueIndex[indexName], fieldM[columnName])
			}
		}
		for indexName, each := range normalKeyMap {
			for _, columnName := range each {
				normalIndex[indexName] = append(normalIndex[indexName], fieldM[columnName])
			}
		}

		checkDuplicateUniqueIndex(uniqueIndex, table.Name)

		list = append(list, &Table{
			Name:        stringx.From(table.Name),
			Db:          stringx.From(database),
			PrimaryKey:  primaryKey,
			UniqueIndex: uniqueIndex,
			Fields:      fields,
		})
	}

	return list, nil
}

func checkDuplicateUniqueIndex(uniqueIndex map[string][]*Field, tableName string) {
	log := console.NewColorConsole()
	uniqueSet := collection.NewSet()
	for k, i := range uniqueIndex {
		var list []string
		for _, field := range i {
			list = append(list, field.Name.Source())
		}

		joinRet := strings.Join(list, ",")
		if uniqueSet.Contains(joinRet) {
			log.Warning("[checkDuplicateUniqueIndex]：表：%s：重复的唯一索引 %s", tableName, joinRet)
			delete(uniqueIndex, k)
			continue
		}

		uniqueSet.AddStr(joinRet)
	}
}

func convertColumns(columns []*parser.Column, primaryColumn string, strict bool) (Primary, map[string]*Field, error) {
	var (
		primaryKey Primary
		fieldM     = make(map[string]*Field)
		log        = console.NewColorConsole()
	)

	for _, column := range columns {
		if column == nil {
			continue
		}

		var (
			comment       string
			isDefaultNull bool
		)

		if column.Constraint != nil {
			comment = column.Constraint.Comment
			isDefaultNull = !column.Constraint.NotNull
			if !column.Constraint.NotNull && column.Constraint.HasDefaultValue {
				isDefaultNull = false
			}

			if column.Name == primaryColumn {
				isDefaultNull = false
			}
		}

		dataType, err := converter.ConvertDataType(column.DataType.Type(), isDefaultNull, column.DataType.Unsigned(), strict)
		if err != nil {
			return Primary{}, nil, err
		}

		if column.Constraint != nil {
			if column.Name == primaryColumn {
				if !column.Constraint.AutoIncrement && dataType == "int64" {
					log.Warning("[convertColumns]：建议主键 %q 添加 `AUTO_INCREMENT` 约束", column.Name)
				}
			} else if column.Constraint.NotNull && !column.Constraint.HasDefaultValue {
				log.Warning("[convertColumns]：建议 %q 列添加 `DEFAULT` 约束", column.Name)
			}
		}

		var field Field
		field.Name = stringx.From(column.Name)
		field.DataType = dataType
		field.Comment = util.TrimNewLine(comment)

		if field.Name.Source() == primaryColumn {
			primaryKey = Primary{
				Field: field,
			}
			if column.Constraint != nil {
				primaryKey.AutoIncrement = column.Constraint.AutoIncrement
			}
		}

		fieldM[field.Name.Source()] = &field
	}

	return primaryKey, fieldM, nil
}

func parseOriginalName(tables []*parser.Table) (originalNames [][]string) {
	var columns []string

	for _, t := range tables {
		columns = []string{}
		for _, c := range t.Columns {
			columns = append(columns, c.Name)
		}
		originalNames = append(originalNames, columns)
	}

	return
}

func getTableFields(table *model.Table, strict bool) (map[string]*Field, error) {
	fieldM := make(map[string]*Field)
	for _, col := range table.Columns {
		isDefaultNull := col.ColumnDefault == nil && col.IsNullAble == "YES"
		isPrimaryUnsigned := strings.Contains(col.ColumnType, "unsigned")
		dt, containsPQ, err := converter.ConvertStringDataType(col.DataType, isDefaultNull, isPrimaryUnsigned, strict)
		if err != nil {
			return nil, err
		}

		colSeqInIndex := 0
		if col.Index != nil {
			colSeqInIndex = col.Index.SeqInIndex
		}

		field := &Field{
			OriginalName:    col.Name,
			Name:            stringx.From(col.Name),
			DataType:        dt,
			Comment:         col.Comment,
			SeqInIndex:      colSeqInIndex,
			OrdinalPosition: col.OrdinalPosition,
			ContainsPQ:      containsPQ,
		}
		fieldM[col.Name] = field
	}

	return fieldM, nil
}
