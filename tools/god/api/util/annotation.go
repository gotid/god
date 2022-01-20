package util

import (
	"strings"

	"github.com/gotid/god/tools/god/api/spec"
)

func GetAnnotationValue(annos []spec.Annotation, key, field string) (string, bool) {
	for _, anno := range annos {
		if anno.Name == field && len(anno.Value) > 0 {
			return anno.Value, true
		}
		if anno.Name == key {
			value, ok := anno.Properties[field]
			return strings.TrimSpace(value), ok
		}
	}
	return "", false
}
