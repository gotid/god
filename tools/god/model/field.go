package model

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"git.zc0901.com/go/god/lib/container/garray"
	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/gconv"
	"git.zc0901.com/go/god/lib/gutil"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/mathx"
	"git.zc0901.com/go/god/lib/mr"
	"git.zc0901.com/go/god/lib/store/cache"
	"git.zc0901.com/go/god/lib/store/sqlx"
	"git.zc0901.com/go/god/lib/stringx"
	"git.zc0901.com/go/god/tools/god/mysql/builder"
)

var (
	fieldFieldList             = builder.FieldList(&Field{})
	fieldFields                = strings.Join(fieldFieldList, ",")
	fieldFieldsAutoSet         = strings.Join(stringx.RemoveDBFields(fieldFieldList, "id", "created_at", "updated_at", "create_time", "update_time"), ",")
	fieldFieldsWithPlaceHolder = strings.Join(stringx.RemoveDBFields(fieldFieldList, "id", "created_at", "updated_at", "create_time", "update_time"), "=?,") + "=?"

	cacheDhomeProjectFieldIdPrefix       = "cache:dhomeProject:field:id:"
	cacheDhomeProjectFieldTableKeyPrefix = "cache:dhomeProject:field:tableKey:"
)

type (
	Field struct {
		Id       int64  `db:"id" json:"Id"`
		TableKey string `db:"table_key" json:"TableKey"` // table:id:virtual_field
		Value    string `db:"value" json:"Value"`        // 支持文本、自定义JSON
	}

	FieldModel struct {
		sqlx.CachedConn
		table string
	}
)

func NewFieldModel(conn sqlx.Conn, clusterConf cache.ClusterConf) *FieldModel {
	return &FieldModel{
		CachedConn: sqlx.NewCachedConnWithCluster(conn, clusterConf),
		table:      "field",
	}
}

func (m *FieldModel) Insert(data Field) (sql.Result, error) {
	query := `insert into ` + m.table + ` (` + fieldFieldsAutoSet + `) values (?, ?)`
	return m.ExecNoCache(query, data.TableKey, data.Value)
}

func (m *FieldModel) TxInsert(tx sqlx.TxSession, data Field) (sql.Result, error) {
	query := `insert into ` + m.table + ` (` + fieldFieldsAutoSet + `) values (?, ?)`
	return tx.Exec(query, data.TableKey, data.Value)
}

func (m *FieldModel) FindOne(id int64) (*Field, error) {
	fieldIdKey := fmt.Sprintf("%s%v", cacheDhomeProjectFieldIdPrefix, id)
	var dest Field
	err := m.Query(&dest, fieldIdKey, func(conn sqlx.Conn, v interface{}) error {
		query := `select ` + fieldFields + ` from ` + m.table + ` where id = ? limit 1`
		return conn.Query(v, query, id)
	})
	if err == nil {
		return &dest, nil
	} else if err == sqlx.ErrNotFound {
		return nil, ErrNotFound
	} else {
		return nil, err
	}
}

func (m *FieldModel) FindMany(ids []int64, workers ...int) (list []*Field) {
	ids = gconv.Int64s(garray.NewArrayFrom(gconv.Interfaces(ids), true).Unique())

	var nWorkers int
	if len(workers) > 0 {
		nWorkers = workers[0]
	} else {
		nWorkers = mathx.MinInt(10, len(ids))
	}

	channel := mr.Map(func(source chan<- interface{}) {
		for _, id := range ids {
			source <- id
		}
	}, func(item interface{}, writer mr.Writer) {
		id := item.(int64)
		one, err := m.FindOne(id)
		if err == nil {
			writer.Write(one)
		} else {
			logx.Error(err)
		}
	}, mr.WithWorkers(nWorkers))

	for one := range channel {
		list = append(list, one.(*Field))
	}

	sort.Slice(list, func(i, j int) bool {
		return gutil.IndexOf(list[i].Id, ids) < gutil.IndexOf(list[j].Id, ids)
	})

	return
}

func (m *FieldModel) FindOneByTableKey(tableKey string) (*Field, error) {
	fieldTableKeyKey := fmt.Sprintf("%s%v", cacheDhomeProjectFieldTableKeyPrefix, tableKey)
	var dest Field
	err := m.QueryIndex(&dest, fieldTableKeyKey, func(primary interface{}) string {
		// 主键的缓存键
		return fmt.Sprintf("%s%v", cacheDhomeProjectFieldIdPrefix, primary)
	}, func(conn sqlx.Conn, v interface{}) (i interface{}, e error) {
		// 无索引建——主键对应缓存，通过索引键查目标行
		query := `select ` + fieldFields + ` from ` + m.table + ` where table_key = ? limit 1`
		if err := conn.Query(&dest, query, tableKey); err != nil {
			return nil, err
		}
		return dest.Id, nil
	}, func(conn sqlx.Conn, v, primary interface{}) error {
		// 如果有索引建——主键对应缓存，则通过主键直接查目标航
		query := `select ` + fieldFields + ` from ` + m.table + ` where id = ? limit 1`
		return conn.Query(v, query, primary)
	})
	if err == nil {
		return &dest, nil
	} else if err == sqlx.ErrNotFound {
		return nil, ErrNotFound
	} else {
		return nil, err
	}
}

func (m *FieldModel) FindManyByTableKeys(keys []string, workers ...int) (list []*Field) {
	keys = gconv.Strings(garray.NewArrayFrom(gconv.Interfaces(keys), true).Unique())

	var nWorkers int
	if len(workers) > 0 {
		nWorkers = workers[0]
	} else {
		nWorkers = mathx.MinInt(10, len(keys))
	}

	channel := mr.Map(func(source chan<- interface{}) {
		for _, key := range keys {
			source <- key
		}
	}, func(item interface{}, writer mr.Writer) {
		key := item.(string)
		one, err := m.FindOneByTableKey(key)
		if err == nil {
			writer.Write(one)
		} else {
			logx.Error(err)
		}
	}, mr.WithWorkers(nWorkers))

	for one := range channel {
		list = append(list, one.(*Field))
	}

	sort.Slice(list, func(i, j int) bool {
		return gutil.IndexOf(list[i].TableKey, keys) < gutil.IndexOf(list[j].TableKey, keys)
	})

	return
}

func (m *FieldModel) Update(data Field) error {
	fieldIdKey := fmt.Sprintf("%s%v", cacheDhomeProjectFieldIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := `update ` + m.table + ` set ` + fieldFieldsWithPlaceHolder + ` where id = ?`
		return conn.Exec(query, data.TableKey, data.Value, data.Id)
	}, fieldIdKey)
	return err
}

func (m *FieldModel) UpdatePartial(data g.Map) error {
	updateArgs, err := sqlx.ExtractUpdateArgs(fieldFieldList, data)
	if err != nil {
		return err
	}

	fieldIdKey := fmt.Sprintf("%s%v", cacheDhomeProjectFieldIdPrefix, updateArgs.Id)
	_, err = m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := `update ` + m.table + ` set ` + updateArgs.Fields + ` where id = ` + updateArgs.Id
		return conn.Exec(query, updateArgs.Args...)
	}, fieldIdKey)
	return err
}

func (m *FieldModel) TxUpdate(tx sqlx.TxSession, data Field) error {
	fieldIdKey := fmt.Sprintf("%s%v", cacheDhomeProjectFieldIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := `update ` + m.table + ` set ` + fieldFieldsWithPlaceHolder + ` where id = ?`
		return tx.Exec(query, data.TableKey, data.Value, data.Id)
	}, fieldIdKey)
	return err
}

func (m *FieldModel) TxUpdatePartial(tx sqlx.TxSession, data g.Map) error {
	updateArgs, err := sqlx.ExtractUpdateArgs(fieldFieldList, data)
	if err != nil {
		return err
	}

	fieldIdKey := fmt.Sprintf("%s%v", cacheDhomeProjectFieldIdPrefix, updateArgs.Id)
	_, err = m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := `update ` + m.table + ` set ` + updateArgs.Fields + ` where id = ` + updateArgs.Id
		return tx.Exec(query, updateArgs.Args...)
	}, fieldIdKey)
	return err
}

func (m *FieldModel) Delete(id ...int64) error {
	if len(id) == 0 {
		return nil
	}

	datas := m.FindMany(id)
	keys := make([]string, len(id)*2)
	for i, v := range id {
		data := datas[i]
		keys[i] = fmt.Sprintf("%s%v", cacheDhomeProjectFieldIdPrefix, v)
		keys[i+1] = fmt.Sprintf("%s%v", cacheDhomeProjectFieldTableKeyPrefix, data.TableKey)
	}

	_, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := fmt.Sprintf(`delete from `+m.table+` where id in (%s)`, sqlx.In(len(id)))
		return conn.Exec(query, gconv.Interfaces(id)...)
	}, keys...)
	return err
}

func (m *FieldModel) TxDelete(tx sqlx.TxSession, id ...int64) error {
	if len(id) == 0 {
		return nil
	}

	datas := m.FindMany(id)
	keys := make([]string, len(id)*2)
	for i, v := range id {
		data := datas[i]
		keys[i] = fmt.Sprintf("%s%v", cacheDhomeProjectFieldIdPrefix, v)
		keys[i+1] = fmt.Sprintf("%s%v", cacheDhomeProjectFieldTableKeyPrefix, data.TableKey)
	}

	_, err := m.Exec(func(conn sqlx.Conn) (result sql.Result, err error) {
		query := fmt.Sprintf(`delete from `+m.table+` where id in (%s)`, sqlx.In(len(id)))
		return tx.Exec(query, gconv.Interfaces(id)...)
	}, keys...)
	return err
}
