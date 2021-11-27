package neo

import (
	"reflect"
	"strings"

	"git.zc0901.com/go/god/lib/breaker"
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
func (d *driver) Driver() neo4j.Driver {
	return d.driver
}

// BeginTx 返回一个新的事务。
func (d *driver) BeginTx() (neo4j.Transaction, error) {
	session := d.driver.NewSession(neo4j.SessionConfig{})
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
	defer func(tx neo4j.Transaction) {
		if tx == nil {
			return
		}
		err := tx.Close()
		if err != nil {
			logx.Errorf("事务关闭失败! %v", err)
		}
	}(tx)

	return d.brk.DoWithAcceptable(func() error {
		return doTx(tx, fn)
	}, d.acceptable)
}

// Read 读数 —— 运行指定 Cypher 并读数至目标。
func (d *driver) Read(ctx Context, dest interface{}, cypher string) error {
	var scanError error
	return d.brk.DoWithAcceptable(func() error {
		ctx.Driver = d.driver
		return doRun(ctx, func(result neo4j.Result) error {
			scanError = scan(dest, result)
			return scanError
		}, cypher)
	}, func(reqError error) bool {
		return reqError == scanError || d.acceptable(reqError)
	})
}

// Run 运行 —— 并利用扫描器扫描指定Cypher的执行结果。
func (d *driver) Run(ctx Context, scanner Scanner, cypher string) error {
	return d.brk.DoWithAcceptable(func() error {
		ctx.Driver = d.driver
		return doRun(ctx, scanner, cypher)
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

func scan(dest interface{}, result neo4j.Result) error {
	// 目标必须为指针类型
	dv := reflect.ValueOf(dest)
	if err := mapping.ValidatePtr(&dv); err != nil {
		return err
	}

	dte := reflect.TypeOf(dest).Elem()
	dve := dv.Elem()
	kind := dte.Kind()
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		err := setOneSimple(dve, result, dv)
		if err != nil {
			return err
		}
	case reflect.Struct:
		err := setOneStruct(dve, result)
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

func setOneStruct(dve reflect.Value, result neo4j.Result) error {
	record, err := result.Single()
	if err != nil {
		return err
	}
	if err := setFieldValueMap(dve, record); err != nil {
		return err
	}

	//for result.Next() {
	//	record := result.Record()
	//
	//	if err := setFieldValueMap(dve, record); err != nil {
	//		return err
	//	}
	//
	//	break
	//}

	return nil
}

func setOneSimple(dve reflect.Value, result neo4j.Result, dv reflect.Value) error {
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
