package main

import (
	"errors"
	"fmt"
	"lambdacalc/shared"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/i582/cfmt/cmd/cfmt"
	"github.com/peterh/liner"
)

// Loading Config into shared Conf variable and starting REPL.
func main() {
	if err := loadConfig(); err != nil {
		return
	}
	cmdline()
}

// Loads config from
// Linux: ".config/labdacalc/config.toml" or Windows: "%APPDATA%/lambda-calc/config.toml"
// If it is not able to do so it loads default config.
func loadConfig() error {
	path := ""
	switch runtime.GOOS {
	case "linux":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(homeDir, ".config", "lambda-calc")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			cfmt.Printf("{{Error:}}red|bold Unable to locat config file. APPDATA not set.")
			return errors.New("no appdata")
		}
		path = filepath.Join(appData, "lambda-calc")
	default:
		cfmt.Println("{{Error:}}red|bold Unsuspected OS. I don't know how to find config file.")
		shared.Conf = shared.GetDefualtConfig()
		return nil
	}

	path = filepath.Join(path, "config.toml")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfmt.Printf("config file not found: %s\n", path)

		if err := createConfig(path); err != nil {
			return err
		}

		shared.Conf = shared.GetDefualtConfig()
		return nil
	}

	if _, err := toml.DecodeFile(path, &shared.Conf); err != nil {
		cfmt.Println("{{Error:}}::red|bold unable to load config file:\n", err)
		shared.Conf = shared.GetDefualtConfig()
		return nil
	}

	return nil
}

func createConfig(path string) error {
	cfmt.Printf("Creating new config file at: %s\n", path)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(shared.GetDefualtConfig()); err != nil {
		return fmt.Errorf("failed to encode default config: %w", err)
	}

	return nil
}

// REPL
func cmdline() {
	var (
		historyFn = filepath.Join(os.TempDir(), ".liner_example_history")
		names     = []string{"define", "solve", "clear", "exit", "help", "drop"}
	)

	line := liner.NewLiner()
	defer line.Close()

	line.SetCtrlCAborts(true)

	line.SetCompleter(func(line string) (c []string) {
		for _, n := range names {
			if strings.HasPrefix(n, strings.ToLower(line)) {
				c = append(c, n)
			}
		}
		return
	})

	if f, err := os.Open(historyFn); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	for {
		arrow := ""
		if shared.Conf.Options["nerdfont"] {
			cfmt.Println("{{}}::blue{{󰪚 Math}}::bgBlue|white{{}}::blue ")
			arrow = "  ╰─▶ "
		} else {
			cfmt.Println("Math")
			arrow = ">>> "
		}
		if cmd, err := line.Prompt(arrow); err == nil {
			line.AppendHistory(cmd)
			switch cmd {
			case "exit":
				cfmt.Println("{{Aborted:}}::yellow|bold Exiting...")
				return
			case "clear":
				clear()
			case "help":
				cfmt.Printf(
					`{{lambda-calc}}::cyan|bold | CLI
Version: 0.0.1

Available commands:

help 		show a help screen with useful information.
exit 		exit the CLI.
define x = ...	define a variable with the value of the equation.
drop x 		undefine a variable.
list  	  list all currently defined shared.Variables.
solve 		solve an equation by a variable if possible.

`)
			default:
				res, err := read(cmd)
				if err == nil {
					cfmt.Printf("%v\n", res)
				}
			}
		} else if err == liner.ErrPromptAborted {
			cfmt.Println("{{Aborted:}}::yellow|bold Exiting...")
			break
		} else {
			cfmt.Println("{{Error:}}::yellow|bold unable to process input.")
		}
	}

	if f, err := os.Create(historyFn); err != nil {
		cfmt.Println("{{Error writing history file:}}::red|bold ", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}
}

func clear() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("clear") // Linux example, its tested
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cfmt.Println("{{Error:}}yellow|bold OS does not support clear command.")
		return
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}
