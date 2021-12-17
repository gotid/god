package translator

import (
	"github.com/pemistahl/lingua-go"
)

// Translator 表示一个翻译器。
type Translator interface {
	// Time 是否开启计时
	Time(b bool)
	// Zh2En 中译英
	Zh2En(query string) string
	// En2Zh 英译中
	En2Zh(query string) string
	// ToZh 自动转成中文
	ToZh(query string) string
	// ToEn 自动转为英文
	ToEn(query string) string
	// Detect 查明语言
	Detect(query string) string
	// Detects 查明语言
	Detects(query string) []lingua.ConfidenceValue
}

type DefaultTranslator struct{}

var languages = []lingua.Language{
	lingua.Chinese,
	lingua.Japanese,
	lingua.Korean,
	lingua.English,
	lingua.French,
	lingua.German,
	lingua.Spanish,
	lingua.Italian,
	lingua.Russian,
	lingua.Ukrainian,
}

var ValidCodes = []string{
	"zh",
	"jp",
	"kor",
	"en",
	"fra",
	"de",
	"spa",
	"it",
	"ru",
	"ukr",
}

var langCodeMap = map[lingua.Language]string{
	lingua.Chinese:   "zh",
	lingua.Japanese:  "jp",
	lingua.Korean:    "kor",
	lingua.English:   "en",
	lingua.French:    "fra",
	lingua.German:    "de",
	lingua.Spanish:   "spa",
	lingua.Italian:   "it",
	lingua.Russian:   "ru",
	lingua.Ukrainian: "ukr",
}

var detector lingua.LanguageDetector

func init() {
	detector = lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		Build()
}

func (d DefaultTranslator) Detect(query string) string {
	if lang, exists := detector.DetectLanguageOf(query); exists {
		return langCodeMap[lang]
	}
	return ""
}

func (d DefaultTranslator) Detects(query string) []lingua.ConfidenceValue {
	return detector.ComputeLanguageConfidenceValues(query)
}
