# OrionLang

Welcome to OrionLang. A general pupose lanaguage inspired by _MonkeyLang_ built in the _Writing An Interpreter In Go_.

## Usage

### REPL

OrionLang supports Read-Evaluate-Print-Loop (REPL). To run the OrionLang REPL run the following:

```sh
go run ./cmd/orionlang/main.go -repl
```

You should see something along the lines of:

```
Hello user! This is the Orion programming language!
Feel free to type in commands
>>
```

### Executing files

To execute files, the OrionLang files must end in `.or`.

Once your file is ready to be executed run the following command:

```sh
go run ./cmd/orionlang/main.go -path {{ PATH_TO_MKL_FILE }}
```

## Features of OrionLang

OrionLang supports the following features:

- Integers
- Booleans
- Strings
- Arrays
- Hashes
- Prefix-, infix- and index operators
- conditionals
- global and local bindings
- first-class functions
- return statements
- closures

### Built-ins

OrionLang supports the following built-in functions:

#### len

Returns the length of an array or string.

```
let x = [1, 2, 3]
puts(len(x)) // 3
```

#### puts

Prints out data to the standard output

```
puts("Hello") // Hello
```

#### first

Returns the first element of an array

```
let x = [1, 2, 3]
puts(first(x)) // 1
```

#### last

Returns the last element of an array

```
let x = [1, 2, 3]
puts(last(x)) // 1
```

#### rest

Returns a new array containing everything after the first element

```
let x = [1, 2, 3]
puts(rest(x)) // [2, 3]
```

#### push

Pushes an element to an array. Returning a new array.

```
let x = [1, 2, 3]
let y = push(x, 4)
puts(x) // [1, 2, 3]
puts(y) // [1, 2, 3, 4]
```
