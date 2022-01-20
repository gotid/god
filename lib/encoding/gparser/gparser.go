// Package gparser provides convenient API for accessing/converting variable and JSON/XML/YAML/TOML.
package gparser

import (
	"github.com/gotid/god/lib/encoding/gjson"
)

// Parser is actually alias of gjson.Json.
type Parser = gjson.Json
