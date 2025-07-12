/*
Package lambdacalc implements a calculation interface based on symbolic 
expressions and mathmatical simplification.
*/
package lambdacalc

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Options map[string]bool
	Symbols map[string]string
	Constants map[string]float64
	// Functions map[string]string
}

var config Config

// Loading Config
func start() bool {
	if _, err := toml.DecodeFile("config/config.toml", &config); err != nil {
		return false
	}
	return true
}
