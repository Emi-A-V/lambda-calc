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
var eventVariableDefinedCallback func(variable Variable)
var eventVariableDroppedCallback func(variable Variable)

// Initilization Parameters:
//
// "VariableDefinedCallback": Callback function called when a variable is defined, with Variable being the parameter.
//
// "VariableDroppedCallback": Callback function called when a variable is dropped, with Variable being the parameter.
type InitilizationParameters struct {
	VariableDefinedCallback func(variable Variable)
	VariableDroppedCallback func(variable Variable)
}

// Loads the config and initilizes the callback functions.
func Start(
	initPrm InitilizationParameters,
) bool {
	eventVariableDefinedCallback = initPrm.VariableDefinedCallback
	eventVariableDroppedCallback = initPrm.VariableDroppedCallback

	if _, err := toml.DecodeFile("config/config.toml", &config); err != nil {
		return false
	}
	return true
}
