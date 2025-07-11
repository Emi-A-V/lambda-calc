package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/i582/cfmt/cmd/cfmt"
	"github.com/peterh/liner"
	// "github.com/alexflint/go-arg"
)

type Config struct {
	Options map[string]bool
	Symbols map[string]string
	Constants map[string]float64
	// Functions map[string]string
}

var config Config

// Loading Config and Parsing Args
func main() {
	// var args struct {
	//	 help bool `arg:"-f --file" help:"Provide a list of all commands"`
	// }
	// arg.MustParse(&args)

	if _, err := toml.DecodeFile("config/config.toml", &config); err != nil {
		cfmt.Println("{{Error:}}::red|bold unable to load config:\n", err)
		return
	}

	cmdline()
}

func cmdline() {
	var (
		historyFn = filepath.Join(os.TempDir(), ".liner_example_history")
		names      = []string{"define", "solve", "clear", "exit", "help", "drop"}
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
		if config.Options["nerdfont"] {
			cfmt.Println("{{}}::blue{{󰪚 Math}}::bgBlue|white{{}}::blue ")
		} else {
			cfmt.Println("{{}}::blue{{Math}}::bgBlue|white{{}}::blue ")
		}
		if cmd, err := line.Prompt("  ╰─▶ "); err == nil {
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
solve 		solve an equation by a variable if possible.

`)
			default:
				res, err := read(cmd)
				if err == nil {	
					cfmt.Printf("%v\n",res)
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
