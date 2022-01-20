package gvalid

import (
	"fmt"

	"github.com/gotid/god/lib/gi18n"
)

// defaultMessages is the default error messages.
// Note that these messages are synchronized from ./i18n/cn/validation.toml .
var defaultMessages = map[string]string{
	"required":             ":attribute 字段不能为空",
	"required-if":          ":attribute 字段不能为空",
	"required-unless":      ":attribute 字段不能为空",
	"required-with":        ":attribute 字段不能为空",
	"required-with-all":    ":attribute 字段不能为空",
	"required-without":     ":attribute 字段不能为空",
	"required-without-all": ":attribute 字段不能为空",
	"date":                 ":attribute 日期格式不正确",
	"date-format":          ":attribute 日期格式不满足:format",
	"email":                ":attribute 邮箱地址格式不正确",
	"phone":                ":attribute 手机号码格式不正确",
	"phone-loose":          ":attribute 手机号码格式不正确",
	"telephone":            ":attribute 电话号码格式不正确",
	"passport":             ":attribute 账号格式不合法，必需以字母开头，只能包含字母、数字和下划线，长度在6~18之间",
	"password":             ":attribute 密码格式不合法，密码格式为任意6-18位的可见字符",
	"password2":            ":attribute 密码格式不合法，密码格式为任意6-18位的可见字符，必须包含大小写字母和数字",
	"password3":            ":attribute 密码格式不合法，密码格式为任意6-18位的可见字符，必须包含大小写字母、数字和特殊字符",
	"postcode":             ":attribute 邮政编码不正确",
	"resident-id":          ":attribute 身份证号码格式不正确",
	"bank-card":            ":attribute 银行卡号格式不正确",
	"qq":                   ":attribute QQ号码格式不正确",
	"ip":                   ":attribute IP地址格式不正确",
	"ipv4":                 ":attribute IPv4地址格式不正确",
	"ipv6":                 ":attribute IPv6地址格式不正确",
	"mac":                  ":attribute MAC地址格式不正确",
	"url":                  ":attribute URL地址格式不正确",
	"domain":               ":attribute 域名格式不正确",
	"length":               ":attribute 字段长度为:min到:max个字符",
	"min-length":           ":attribute 字段最小长度为:min",
	"max-length":           ":attribute 字段最大长度为:max",
	"between":              ":attribute 字段大小为:min到:max",
	"min":                  ":attribute 字段最小值为:min",
	"max":                  ":attribute 字段最大值为:max",
	"json":                 ":attribute 字段应当为JSON格式",
	"xml":                  ":attribute 字段应当为XML格式",
	"array":                ":attribute 字段应当为数组",
	"integer":              ":attribute 字段应当为整数",
	"float":                ":attribute 字段应当为浮点数",
	"boolean":              ":attribute 字段应当为布尔值",
	"same":                 ":attribute 字段值必须和:field相同",
	"different":            ":attribute 字段值不能与:field相同",
	"in":                   ":attribute 字段值不合法",
	"not-in":               ":attribute 字段值不合法",
	"regex":                ":attribute 字段值不合法",
	"__default__":          ":attribute 字段值不合法",
}

// getErrorMessageByRule retrieves and returns the error message for specified rule.
// It firstly retrieves the message from custom message map, and then checks i18n manager,
// it returns the default error message if it's not found in custom message map or i18n manager.
func getErrorMessageByRule(ruleKey string, customMsgMap map[string]string) string {
	content := customMsgMap[ruleKey]
	if content != "" {
		return content
	}
	content = gi18n.GetContent(fmt.Sprintf(`god.gvalid.rule.%s`, ruleKey))
	if content == "" {
		content = defaultMessages[ruleKey]
	}
	// If there's no configured rule message, it uses default one.
	if content == "" {
		content = gi18n.GetContent(`god.gvalid.rule.__default__`)
		if content == "" {
			content = defaultMessages["__default__"]
		}
	}
	return content
}
