package neo

import (
	"errors"
	"reflect"
	"strings"

	"git.zc0901.com/go/god/lib/breaker"
	"git.zc0901.com/go/god/lib/g"
	"git.zc0901.com/go/god/lib/logx"
	"git.zc0901.com/go/god/lib/mapping"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const (
	tagName = "neo"
)

var (
	ErrNotReadableValue     = errors.New("neo: 无法读取的值，检查结构字段是否大写开头")
	ErrUnsupportedValueType = errors.New("neo: 不支持的扫描目标类型")
)

type (
	// Session 表示一个可进行 neo4j 读写的会话。
	Session interface {
		// Read 读数 —— 运行Cypher并读入目标。
		Read(dest interface{}, cypher string, params ...g.Map) error
		// TxRead 事务型读数 —— 运行Cypher并读入目标。
		TxRead(tx neo4j.Transaction, dest interface{}, cypher string, params ...g.Map) error
		// Scan 扫数 —— 利用扫描器扫描指定Cypher的查询结果。
		Scan(scanner Scanner, cypher string, params ...g.Map) error
		// TxScan 事务型扫数 —— 利用扫描器扫描指定Cypher的查询结果。
		TxScan(tx neo4j.Transaction, scanner Scanner, cypher string, params ...g.Map) error
	}

	// Driver 表示一个带有断路器保护的 neo4j 驱动。
	Driver interface {
		Session
	}

	driver struct {
		target,
		username,
		password,
		realm string // neo4j.Driver 连接字符串
		brk    breaker.Breaker         // 断路器
		accept func(reqErr error) bool // 自定义错误可接收器
	}
)

var _ Driver = (*driver)(nil)

// NewDriver 返回一个新的 neo4j 驱动。
func NewDriver(target, username, password, realm string) Driver {
	d := &driver{
		target:   target,
		username: username,
		password: password,
		realm:    realm,
		brk:      breaker.NewBreaker(),
	}
	return d
}

// Read 读数 —— 运行指定 Cypher 并读数至目标。
func (d *driver) Read(dest interface{}, cypher string, params ...g.Map) error {
	var readError error
	err := d.brk.DoWithAcceptable(func() error {
		driver4j, err := getDriver(d.target, d.username, d.password, d.realm)
		if err != nil {
			logConnError(d.target, err)
			return err
		}

		err = doRun(driver4j, func(result neo4j.Result) error {
			return scan(dest, result)
		}, cypher, params...)
		return err
	}, func(reqError error) bool {
		return reqError == readError || d.acceptable(reqError)
	})
	if err != nil {
		return err
	}

	return nil
}

// TxRead 事务型读数 —— 运行指定 Cypher 并读数至目标。
func (d *driver) TxRead(tx neo4j.Transaction, dest interface{}, cypher string, params ...g.Map) error {
	var readError error
	err := d.brk.DoWithAcceptable(func() error {
		err := doTxRun(tx, func(result neo4j.Result) error {
			return scan(dest, result)
		}, cypher, params...)
		return err
	}, func(reqError error) bool {
		return reqError == readError || d.acceptable(reqError)
	})
	if err != nil {
		return err
	}

	return nil
}

// Scan 扫数 —— 运行指定 Cypher 查询并利用指定扫描器进行结果扫描。
func (d *driver) Scan(scanner Scanner, cypher string, params ...g.Map) error {
	var readError error
	err := d.brk.DoWithAcceptable(func() error {
		driver4j, err := getDriver(d.target, d.username, d.password, d.realm)
		if err != nil {
			logConnError(d.target, err)
			return err
		}

		err = doRun(driver4j, scanner, cypher, params...)
		return err
	}, func(reqError error) bool {
		return reqError == readError || d.acceptable(reqError)
	})
	if err != nil {
		return err
	}

	return nil
}

// TxScan 事务型扫数 —— 运行指定 Cypher 查询并利用指定扫描器进行结果扫描。
func (d *driver) TxScan(tx neo4j.Transaction, scanner Scanner, cypher string, params ...g.Map) error {
	var readError error
	err := d.brk.DoWithAcceptable(func() error {
		err := doTxRun(tx, scanner, cypher, params...)
		return err
	}, func(reqError error) bool {
		return reqError == readError || d.acceptable(reqError)
	})
	if err != nil {
		return err
	}

	return nil
}

// 判断错误是否可接受
func (d *driver) acceptable(reqError error) bool {
	ok := reqError == nil
	if d.accept == nil {
		return ok
	}
	return ok || d.accept(reqError)
}

func logConnError(target string, err error) {
	logx.Errorf("获取 neo4j 连接池失败 %s: %v", target, err)
}

func scan(dest interface{}, result neo4j.Result) error {
	// 目标必须为指针类型
	dv := reflect.ValueOf(dest)
	if err := mapping.ValidatePtr(&dv); err != nil {
		return err
	}

	dte := reflect.TypeOf(dest).Elem()
	dve := dv.Elem()
	switch dte.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		err := setSimpleValue(dve, result, dv)
		if err != nil {
			return err
		}
	case reflect.Struct:
		err := setStructValue(dve, result)
		if err != nil {
			return err
		}
	case reflect.Slice:
		if !dve.CanSet() {
			return ErrNotSettable
		}
		ptr := dte.Elem().Kind() == reflect.Ptr
		appendFn := func(item reflect.Value) {
			if ptr {
				dve.Set(reflect.Append(dve, item))
			} else {
				dve.Set(reflect.Append(dve, reflect.Indirect(item)))
			}
		}
		fillFn := func(field, value interface{}) error {
			if dve.CanSet() {
				f := reflect.Indirect(reflect.ValueOf(field))
				f.Set(reflect.ValueOf(value))
				appendFn(reflect.ValueOf(field))
				return nil
			}
			return ErrNotSettable
		}

		base := mapping.Deref(dte.Elem())
		baseKind := base.Kind()
		switch baseKind {
		case reflect.String, reflect.Bool, reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			for result.Next() {
				field := reflect.New(base)
				if err := fillFn(field.Interface(), result.Record().Values[0]); err != nil {
					return err
				}
			}
		case reflect.Struct:
			for result.Next() {
				record := result.Record()
				structValue := reflect.New(base)
				if err := setFieldValueMap(structValue, record); err != nil {
					return err
				}
				appendFn(structValue)
			}
			return nil
		default:
			return ErrUnsupportedValueType
		}
		return nil
	default:
		return ErrUnsupportedValueType
	}

	return nil
}

func setStructValue(dve reflect.Value, result neo4j.Result) error {
	for result.Next() {
		record := result.Record()
		if err := setFieldValueMap(dve, record); err != nil {
			return err
		}
		break
	}
	return nil
}

func setSimpleValue(dve reflect.Value, result neo4j.Result, dv reflect.Value) error {
	if dve.CanSet() {
		for result.Next() {
			record := result.Record()
			reflect.Indirect(dv).Set(reflect.ValueOf(record.Values[0]))
			return nil
		}
	} else {
		return ErrNotSettable
	}
	return nil
}

// setFieldValueMap: 获取结构体字段中标记的字段名——值映射关系
// 在编写字段tag的情况下，可以确保结构体字段和Cypher选择列不一致的情况下不出错
func setFieldValueMap(structValue reflect.Value, record *neo4j.Record) error {
	dt := mapping.Deref(structValue.Type())
	size := dt.NumField()

	for i := 0; i < size; i++ {
		// 取字段标记中的列名，如`neo:"total"` 中的 total
		fieldName := getFieldTag(dt.Field(i))
		if len(fieldName) == 0 {
			continue
		}

		// 读取指针字段或非指针字段的值
		field := reflect.Indirect(structValue).Field(i)
		kind := field.Kind()
		switch kind {
		case reflect.Ptr:
			if !field.CanInterface() {
				return ErrNotReadableValue
			}
			if v, ok := record.Get(fieldName); ok {
				field.Set(reflect.ValueOf(v))
			}
		default:
			if !field.CanAddr() || !field.Addr().CanInterface() {
				return ErrNotReadableValue
			}
			if v, ok := record.Get(fieldName); ok {
				field.Set(reflect.ValueOf(v))
			}
		}
	}

	return nil
}

// getFieldTag 获取结构标记值
func getFieldTag(field reflect.StructField) string {
	tag := field.Tag.Get(tagName)
	if len(tag) == 0 {
		return ""
	}
	return strings.Split(tag, ",")[0]
}
