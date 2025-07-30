# lambda-calc (engine)
**L**ogical **A**dvanced **M**athmatical **B**ackend for **D**ynamic **A**nalysis of **C**alculation and **A**lgebra with **L**aTeX **C**uteness

This is a modular version of the lambda-calc project. It is made for the lambda-calc-gui, but should work in any GO-Project. For more information and work in progress features of this project look at the [main](https://github.com/Emi-A-V/lambda-calc/) branch.

It is currently not a go package. So if you want to use it, you will need to but the folder into your project.

## Usage

You can add the lambda-calc-engine into your GO-project and import it.

```go
import (
    lambda "lambda-calc-gui/lambda-calc"
)
```

To initialize the engine you need to call the `start` function. An initilization parameter that holds all callback functions is passed to the function.
```go
func main() {
    lambda.Start(lambda.InitilizationParameters{
        func(lambda.Variable) {} // Callback, when variables are defined.
        func(lambda.Variable) {} // Callback, when variables are dropped.
    })
}
```

The Variable struct passed into these callback function holds information about the variable like its name and its value.
```go 
type Variable struct {
	Name     string
	Equation string
}
```
To now calculate the result of an equation you need to call the `Input` function. The function returns the result of type string and an error. The error contains the whether and error was produced, the error message and an error code.
```go
func main() {
    str := "2 + 2"

    res, err := lambda.Input(str)
    if err.IsError {
        fmt.Printf("Recieved Error: %s", err.Message)
    } else {
        fmt.Printf("Recieved Result: %s", res)
    }
}
```

## Config-File
The config file can be found under `config/config.toml`. The config file defines the behavior of the math engine.

#### Options
Currently options are only there to debug the program and trace back at what point an error occured.
```toml
[options]
show_debug_process = false
```
|Option - _bool_|Effect|
|---------------|------|
|`show_debug_process`|Prints out message about the state of the program during calculation.|

#### Symbols
All symbols used when entering an equation can be configured:
```toml
[symbols]
decimal_split = '.'

plus = '+'
minus = '-'
multiply = '*'
divide = '/'

sqrt = 'sqrt'
power = '^'
```
Currently `sqrt` is the only multi-character symbol.
#### Constants
Constants can also be defined in the config file. All constants are declared under the constant struct, with the key being the name of the variable and the value the value of the variable. The value is in a Float64 format. By default these constants are included:
```toml
[constants]
pi = 3.14159265
phi = 1.618033988
e = 2.71828182
```
