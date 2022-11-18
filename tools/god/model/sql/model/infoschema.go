package model

import (
	"fmt"
	"github.com/gotid/god/lib/store/sqlx"
	"github.com/gotid/god/tools/god/model/sql/util"
	"sort"
)

const indexPrimary = "PRIMARY"

type (
	// InformationSchemaModel 定义信息架构模型。
	InformationSchemaModel struct {
		conn sqlx.Conn
	}

	// DbIndex 定义 information_schema.statistic 表中列的索引。
	DbIndex struct {
		IndexName  string `db:"INDEX_NAME"`
		NonUnique  int    `db:"NON_UNIQUE"`
		SeqInIndex int    `db:"SEQ_IN_INDEX"`
	}

	// DbColumn 定义列信息。
	DbColumn struct {
		Name            string `db:"COLUMN_NAME"`
		DataType        string `db:"DATA_TYPE"`
		ColumnType      string `db:"COLUMN_TYPE"`
		Extra           string `db:"EXTRA"`
		Comment         string `db:"COLUMN_COMMENT"`
		ColumnDefault   any    `db:"COLUMN_DEFAULT"`
		IsNullAble      string `db:"IS_NULLABLE"`
		OrdinalPosition int    `db:"ORDINAL_POSITION"`
	}

	// Column 定义表中的列。
	Column struct {
		*DbColumn
		Index *DbIndex
	}

	// ColumnData 描述表中的列。
	ColumnData struct {
		Db      string
		Table   string
		Columns []*Column
	}

	// IndexType 索引类型，是字符串的别名。
	IndexType string

	// Index 定义索引。
	Index struct {
		IndexType IndexType
		Columns   []*Column
	}

	// Table 定义表。
	Table struct {
		Db          string
		Table       string
		Columns     []*Column
		PrimaryKey  *Column
		UniqueIndex map[string][]*Column
		NormalIndex map[string][]*Column
	}
)

// NewInformationSchemaModel 返回一个新的 InformationSchemaModel 实例。
func NewInformationSchemaModel(conn sqlx.Conn) *InformationSchemaModel {
	return &InformationSchemaModel{
		conn: conn,
	}
}

// GetAllTables 获取给定数据库的所有表格。
func (m *InformationSchemaModel) GetAllTables(database string) ([]string, error) {
	query := `select TABLE_NAME from TABLES where TABLE_SCHEMA = ?`
	var tables []string
	err := m.conn.QueryRows(&tables, query, database)
	if err != nil {
		return nil, err
	}

	return tables, nil
}

// FindColumns 获取给定数据库表的列数据。
func (m *InformationSchemaModel) FindColumns(db, table string) (*ColumnData, error) {
	query := `select c.COLUMN_NAME,c.DATA_TYPE,c.COLUMN_TYPE,EXTRA,c.COLUMN_COMMENT,c.COLUMN_DEFAULT,c.IS_NULLABLE,c.ORDINAL_POSITION from COLUMNS c where c.TABLE_SCHEMA = ? and c.TABLE_NAME = ?`
	var reply []*DbColumn
	err := m.conn.QueryRowsPartial(&reply, query, db, table)
	if err != nil {
		return nil, err
	}

	var columns []*Column
	for _, column := range reply {
		index, err := m.FindIndex(db, table, column.Name)
		if err != nil {
			if err != sqlx.ErrNotFound {
				return nil, err
			}

			continue
		}

		if len(index) > 0 {
			for _, dbIndex := range index {
				columns = append(columns, &Column{
					DbColumn: column,
					Index:    dbIndex,
				})
			}
		} else {
			columns = append(columns, &Column{
				DbColumn: column,
			})
		}
	}

	sort.Slice(columns, func(i, j int) bool {
		return columns[i].OrdinalPosition < columns[j].OrdinalPosition
	})

	columnData := ColumnData{
		Db:      db,
		Table:   table,
		Columns: columns,
	}

	return &columnData, nil
}

// FindIndex 获取给定数据库表中给定列的索引。
func (m *InformationSchemaModel) FindIndex(db, table, column string) ([]*DbIndex, error) {
	query := `select s.INDEX_NAME,s.NON_UNIQUE,s.SEQ_IN_INDEX from  STATISTICS s where s.TABLE_SCHEMA = ? and s.TABLE_NAME = ? and s.COLUMN_NAME = ?`
	var reply []*DbIndex
	err := m.conn.QueryRowsPartial(&reply, query, db, table, column)
	if err != nil {
		return nil, err
	}

	return reply, nil
}

// Convert 将列数据 cd 转为 Table。
func (c *ColumnData) Convert() (*Table, error) {
	table := Table{
		Db:          c.Db,
		Table:       c.Table,
		Columns:     c.Columns,
		PrimaryKey:  nil,
		UniqueIndex: map[string][]*Column{},
		NormalIndex: map[string][]*Column{},
	}

	m := make(map[string][]*Column)
	for _, column := range c.Columns {
		column.Comment = util.TrimNewLine(column.Comment)
		if column.Index != nil {
			m[column.Index.IndexName] = append(m[column.Index.IndexName], column)
		}
	}

	primaryColumns := m[indexPrimary]
	if len(primaryColumns) == 0 {
		return nil, fmt.Errorf("数据库：%s，表：%s，缺失主键", c.Db, c.Table)
	}
	if len(primaryColumns) > 1 {
		return nil, fmt.Errorf("数据库：%s，表：%s，联合主键不被 god 支持", c.Db, c.Table)
	}

	table.PrimaryKey = primaryColumns[0]
	for indexName, columns := range m {
		if indexName == indexPrimary {
			continue
		}

		for _, col := range columns {
			if col.Index != nil {
				if col.Index.NonUnique == 0 {
					table.UniqueIndex[indexName] = columns
				} else {
					table.NormalIndex[indexName] = columns
				}
			}
		}
	}

	return &table, nil
}
