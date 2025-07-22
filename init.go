/*
Package lambdaengine implements a calculation interface based on symbolic
expressions and mathmatical simplification.
*/
package lambdaengine

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Options   map[string]bool
	Symbols   map[string]string
	Constants map[string]float64
	// Functions map[string]string
}

// Global config variable
var config Config

// Global callback-functions.
var eventVariableDefinedCallback func(variable Var)
var eventVariableDroppedCallback func(variable Var)

// Loading Config
func Start(
	varDefCallback func(variable Var),
	varDropCallback func(variable Var),
) bool {
	eventVariableDefinedCallback = varDefCallback
	eventVariableDroppedCallback = varDropCallback

	if _, err := toml.DecodeFile("config/config.toml", &config); err != nil {
		return false
	}
	return true
}
