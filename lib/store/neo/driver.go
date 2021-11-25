package neo

import (
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

// 带有断路器保护的 neo4j 驱动
type driver struct {
	target,
	username,
	password,
	realm string // neo4j.Driver 连接字符串
	driver neo4j.Driver            // neo4j.Driver 驱动
	brk    breaker.Breaker         // 断路器
	accept func(reqErr error) bool // 自定义错误可接收器
}

var _ Driver = (*driver)(nil)

// MustDriver 返回一个新的 neo4j 驱动。
func MustDriver(target, username, password, realm string) Driver {
	d := &driver{
		target:   target,
		username: username,
		password: password,
		realm:    realm,
		brk:      breaker.NewBreaker(),
	}
	driver4j, err := getDriver(d.target, d.username, d.password, d.realm)
	d.driver = driver4j
	if err != nil {
		logx.Errorf("neo.MustDriver 初始化失败！")
		panic(err)
	}

	return d
}

// Driver 返回可复用的 neo4j.Driver。
func (d *driver) Driver() (neo4j.Driver, error) {
	return getDriver(d.target, d.username, d.password, d.realm)
}

// BeginTx 返回一个新的事务。
func (d *driver) BeginTx() (neo4j.Transaction, error) {
	driver4j, err := d.Driver()
	if err != nil {
		return nil, err
	}
	session := driver4j.NewSession(neo4j.SessionConfig{})
	tx, err := session.BeginTransaction()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Transact 执行事务型操作。
func (d *driver) Transact(fn TransactFn) error {
	tx, err := d.BeginTx()
	if err != nil {
		return err
	}

	return d.brk.DoWithAcceptable(func() error {
		return doTx(tx, fn)
	}, d.acceptable)
}

// Read 读数 —— 运行指定 Cypher 并读数至目标。
func (d *driver) Read(dest interface{}, cypher string, params ...g.Map) error {
	var scanError error
	return d.brk.DoWithAcceptable(func() error {
		driver4j, err := d.Driver()
		if err != nil {
			logConnError(d.target, err)
			return err
		}

		return doRun(driver4j, func(result neo4j.Result) error {
			scanError = scan(dest, result)
			return scanError
		}, cypher, params...)
	}, func(reqError error) bool {
		return reqError == scanError || d.acceptable(reqError)
	})
}

// TxRead 事务型读数 —— 运行指定 Cypher 并读数至目标。
func (d *driver) TxRead(tx neo4j.Transaction, dest interface{}, cypher string, params ...g.Map) error {
	var scanError error
	return d.brk.DoWithAcceptable(func() error {
		return doTxRun(tx, func(result neo4j.Result) error {
			scanError = scan(dest, result)
			return scanError
		}, cypher, params...)
	}, func(reqError error) bool {
		return reqError == scanError || d.acceptable(reqError)
	})
}

// Run 运行 —— 并利用扫描器扫描指定Cypher的执行结果。
func (d *driver) Run(scanner Scanner, cypher string, params ...g.Map) error {
	return d.brk.DoWithAcceptable(func() error {
		driver4j, err := d.Driver()
		if err != nil {
			logConnError(d.target, err)
			return err
		}

		return doRun(driver4j, scanner, cypher, params...)
	}, func(reqError error) bool {
		return d.acceptable(reqError)
	})
}

// TxRun 事务型运行 —— 利用扫描器扫描指定Cypher的执行结果。
func (d *driver) TxRun(tx neo4j.Transaction, scanner Scanner, cypher string, params ...g.Map) error {
	return d.brk.DoWithAcceptable(func() error {
		return doTxRun(tx, scanner, cypher, params...)
	}, func(reqError error) bool {
		return d.acceptable(reqError)
	})
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

func getFieldTag(field reflect.StructField) string {
	tag := field.Tag.Get(tagName)
	if len(tag) == 0 {
		return ""
	}
	return strings.Split(tag, ",")[0]
}
