package translate

// Translator 表示一个翻译器。
type Translator interface {
	// Zh2En 中译英
	Zh2En(query string) string
	// En2Zh 英译中
	En2Zh(query string) string
}
