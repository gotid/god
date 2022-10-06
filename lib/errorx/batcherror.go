package errorx

import "bytes"

// BatchError 是一个保存多个错误的结构体。
type BatchError struct {
	errs errorArray
}

// Add 添加一组 errs 到 be，忽略空忽略。
func (be *BatchError) Add(errs ...error) {
	for _, err := range errs {
		if err != nil {
			be.errs = append(be.errs, err)
		}
	}
}

// Err 返回表示所有错误的错误。
func (be *BatchError) Err() error {
	switch len(be.errs) {
	case 0:
		return nil
	case 1:
		return be.errs[0]
	default:
		return be.errs
	}
}

// NotNil 检查内部是否有错误。
func (be *BatchError) NotNil() bool {
	return len(be.errs) > 0
}

type errorArray []error

// 返回一个表示内部错误的字符串。
func (ea errorArray) Error() string {
	var buf bytes.Buffer

	for i := range ea {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(ea[i].Error())
	}

	return buf.String()
}
