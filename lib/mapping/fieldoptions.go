package mapping

import "fmt"

const notSymbol = '!'

type (
	// 用上下文和 OptionalDep 选项来决定 Optional 的值
	// 对于 context.Context 则什么都不做
	fieldOptionsWithContext struct {
		Inherit    bool
		FromString bool
		Optional   bool
		Options    []string
		Default    string
		EnvVar     string
		Range      *numberRange
	}

	fieldOptions struct {
		fieldOptionsWithContext
		OptionalDep string
	}

	numberRange struct {
		left         float64
		leftInclude  bool
		right        float64
		rightInclude bool
	}
)

func (o *fieldOptionsWithContext) fromString() bool {
	return o != nil && o.FromString
}

func (o *fieldOptionsWithContext) getDefault() (string, bool) {
	if o == nil {
		return "", false
	}

	return o.Default, len(o.Default) > 0
}

func (o *fieldOptionsWithContext) inherit() bool {
	return o != nil && o.Inherit
}

func (o *fieldOptionsWithContext) optional() bool {
	return o != nil && o.Optional
}

func (o *fieldOptionsWithContext) options() []string {
	if o == nil {
		return nil
	}

	return o.Options
}

func (o *fieldOptions) optionalDep() string {
	if o == nil {
		return ""
	}

	return o.OptionalDep
}

func (o *fieldOptions) toOptionsWithContext(key string, m Valuer, fullName string) (*fieldOptionsWithContext, error) {
	var optional bool
	if o.optional() {
		dep := o.optionalDep()
		if len(dep) == 0 {
			optional = true
		} else if dep[0] == notSymbol {
			dep = dep[1:]
			if len(dep) == 0 {
				return nil, fmt.Errorf("%q 中 %q 的可选值错误", fullName, key)
			}

			_, baseOn := m.Value(dep)
			_, selfOn := m.Value(key)
			if baseOn == selfOn {
				return nil, fmt.Errorf("%q 中要么设定 %q 要么设定 %q", fullName, dep, key)
			}

			optional = baseOn
		} else {
			_, baseOn := m.Value(dep)
			_, selfOn := m.Value(key)
			if baseOn != selfOn {
				return nil, fmt.Errorf("%q 中 %q 和 %q 要么都设定，要么都别设定", fullName, dep, key)
			}

			optional = !baseOn
		}
	}

	if o.fieldOptionsWithContext.Optional == optional {
		return &o.fieldOptionsWithContext, nil
	}

	return &fieldOptionsWithContext{
		FromString: o.FromString,
		Optional:   optional,
		Options:    o.Options,
		Default:    o.Default,
	}, nil
}
