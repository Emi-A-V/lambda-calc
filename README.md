# lambda-calc

**L**ogical **A**dvanced **M**athmatical **B**ackend for **D**ynamic **A**nalysis of **C**alculation and **A**lgebra with **L**aTeX **C**uteness

lambda-calc is my shot at an advanced scientific calculator with symbolic expressions. It is written in GO, because it like GO :3

### Features
-[x] Simple Number Calculation
-[x] CLI REPL interface
-[x] Defining Variables

### Planned Features
-[ ] Simplifying equations (work in progress)
-[ ] Solving equations
-[ ] Support for reading and writing LaTeX
-[ ] More fancy math features

## Command Line Interface
Entering an equation will result in its calculation, if the calculation is solvable.
```
2 + 2
-> 4
```

You can define a variable with the `define` keyword:
```
define x = 5 * 2 + 3
-> Variable defined.
```
The variable can afterwards be used like a number:
```
2 + x
-> 15
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
