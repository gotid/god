package baidu

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"git.zc0901.com/go/god/lib/translate/translator"

	"git.zc0901.com/go/god/lib/logx"

	"git.zc0901.com/go/god/lib/stringx"

	"github.com/imroc/req"

	v8 "rogchap.com/v8go"
)

const signSource = `function a(r) {
    if (Array.isArray(r)) {
        for (var o = 0, t = Array(r.length); o < r.length; o++)
            t[o] = r[o];
        return t
    }
    return Array.from(r)
}
function n(r, o) {
    for (var t = 0; t < o.length - 2; t += 3) {
        var a = o.charAt(t + 2);
        a = a >= "a" ? a.charCodeAt(0) - 87 : Number(a),
            a = "+" === o.charAt(t + 1) ? r >>> a : r << a,
            r = "+" === o.charAt(t) ? r + a & 4294967295 : r ^ a
    }
    return r
}
function e(r, gtk) {
    var o = r.match(/[\uD800-\uDBFF][\uDC00-\uDFFF]/g);
    if (null === o) {
        var t = r.length;
        t > 30 && (r = "" + r.substr(0, 10) + r.substr(Math.floor(t / 2) - 5, 10) + r.substr(-10, 10))
    } else {
        for (var e = r.split(/[\uD800-\uDBFF][\uDC00-\uDFFF]/), C = 0, h = e.length, f = []; h > C; C++)
            "" !== e[C] && f.push.apply(f, a(e[C].split(""))),
            C !== h - 1 && f.push(o[C]);
        var g = f.length;
        g > 30 && (r = f.slice(0, 10).join("") + f.slice(Math.floor(g / 2) - 5, Math.floor(g / 2) + 5).join("") + f.slice(-10).join(""))
    }
    var u = void 0;
    u = null !== i ? i : (i = gtk || "") || "";
    for (var d = u.split("."), m = Number(d[0]) || 0, s = Number(d[1]) || 0, S = [], c = 0, v = 0; v < r.length; v++) {
        var A = r.charCodeAt(v);
        128 > A ? S[c++] = A : (2048 > A ? S[c++] = A >> 6 | 192 : (55296 === (64512 & A) && v + 1 < r.length && 56320 === (64512 & r.charCodeAt(v + 1)) ? (A = 65536 + ((1023 & A) << 10) + (1023 & r.charCodeAt(++v)),
            S[c++] = A >> 18 | 240,
            S[c++] = A >> 12 & 63 | 128) : S[c++] = A >> 12 | 224,
            S[c++] = A >> 6 & 63 | 128),
            S[c++] = 63 & A | 128)
    }
    for (var p = m, F = "" + String.fromCharCode(43) + String.fromCharCode(45) + String.fromCharCode(97) + ("" + String.fromCharCode(94) + String.fromCharCode(43) + String.fromCharCode(54)), D = "" + String.fromCharCode(43) + String.fromCharCode(45) + String.fromCharCode(51) + ("" + String.fromCharCode(94) + String.fromCharCode(43) + String.fromCharCode(98)) + ("" + String.fromCharCode(43) + String.fromCharCode(45) + String.fromCharCode(102)), b = 0; b < S.length; b++)
        p += S[b],
            p = n(p, F);
    return p = n(p, D),
        p ^= s,
    0 > p && (p = (2147483647 & p) + 2147483648),
        p %= 1e6,
    p.toString() + "." + (p ^ m)
}
var i = null;`

type Translate struct {
	ctx        *v8.Context
	token, gtk string
}

var _ translator.Translator = (*Translate)(nil)

var instance *Translate

var once = new(sync.Once)

var (
	reToken = regexp.MustCompile(`token: '(.*?)'`)
	reGtk   = regexp.MustCompile(`window.gtk = '(.*?)'`)
	reDst   = regexp.MustCompile(`"dst":"(.*?)"`)
)

var header = req.Header{
	"Host":       "fanyi.baidu.com",
	"Referer":    "https://fanyi.baidu.com/",
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.89 Safari/537.36",
}

// Must 返回百度翻译器。
func Must() translator.Translator {
	once.Do(func() {
		ctx := v8.NewContext() // 使用新的虚拟机创建新的V8上下文
		_, err := ctx.RunScript(signSource, "sign.js")
		if err != nil {
			panic(err)
		}
		instance = &Translate{ctx: ctx}
	})
	return instance
}

func (t *Translate) ensureToken() {
	if t.token == "" {
		token, gtk := instance.getTokenAndGtk()
		instance.token = token
		instance.gtk = gtk
	}
}

func (t *Translate) Zh2En(query string) string {
	t.ensureToken()
	return t.translate(query, "zh", "en")
}

func (t *Translate) En2Zh(query string) string {
	t.ensureToken()
	return t.translate(query, "en", "zh")
}

func (t *Translate) Detect(query string) string {
	// TODO implement me
	panic("implement me")
}

func (t *Translate) translate(query, from, to string) string {
	if !stringx.Contains([]string{"zh", "en"}, from) ||
		!stringx.Contains([]string{"zh", "en"}, to) {
		return ""
	}
	sign, err := t.sign(query, t.gtk)
	url := fmt.Sprintf("https://fanyi.baidu.com/v2transapi?from=%s&to=%s", from, to)
	resp, err := req.Post(url, req.Param{
		"from":              from,
		"to":                to,
		"query":             query,
		"transtype":         "enter",
		"simple_means_flag": "3",
		"sign":              sign,
		"token":             t.token,
		"domain":            "common",
	}, header)
	if err != nil {
		logx.Errorf("百度翻译失败: %v", err)
		return ""
	}

	ss := resp.String()
	s := reDst.FindAllStringSubmatch(ss, -1)
	if s == nil {
		logx.Error("百度翻译结果查找失败:", zhToUnicode(ss))

		return ""
	}
	if to == "zh" {
		return zhToUnicode(s[0][1])
	}
	return s[0][1]
}

func (t *Translate) getTokenAndGtk() (token, gtk string) {
	token, gtk = t.tokenAndGtk()
	if token == "" {
		token, gtk = t.tokenAndGtk()
		if token == "" {
			token, gtk = t.tokenAndGtk()
		}
	}
	return token, gtk
}

func (t *Translate) tokenAndGtk() (token, gtk string) {
	resp, err := req.Get("https://fanyi.baidu.com/", header)
	if err != nil {
		return "", ""
	}

	tk := reToken.FindAllStringSubmatch(resp.String(), -1)
	gt := reGtk.FindAllStringSubmatch(resp.String(), -1)
	if tk == nil || gt == nil {
		return token, gtk
	}
	token = tk[0][1]
	gtk = gt[0][1]

	return token, gtk
}

func (t *Translate) sign(query, gtk string) (string, error) {
	_, err := t.ctx.RunScript(signSource, "sign.js")
	if err != nil {
		return "", err
	}

	_, err = t.ctx.RunScript(fmt.Sprintf(`var result=e("%s", "%s")`, query, gtk), "main.js")
	if err != nil {
		return "", err
	}

	val, err := t.ctx.RunScript("result", "value.js")
	return val.String(), nil
}

func zhToUnicode(raw string) string {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(raw), `\\u`, `\u`, -1))
	if err != nil {
		logx.Errorf("百度英译中转码错误：%v", err)
		return ""
	}
	return str
}
