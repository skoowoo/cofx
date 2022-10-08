# Grammar Introduction

## Comment
Use `//` to add code comments. :warning: Note that only exclusive line comments are provided, not support end-of-line comments.

## load
load is used to load a function, for example: load the function 'print'

```go
// go is a function driver, which means that the function print is a piece of Go code and needs to be run by the go driver
// print is the function name
load go:print
```

All functions need to be loaded before they can be used.

## var
The `var` keyword can define a variable, :warning: Note: The variable itself has no type, but the built-in default distinguishes between strings and numbers, and numeric variables can perform arithmetic operations.

```go
var a = "Hello World!"
// $(a) means to get the value of the variable a
var b = $(a)
var c = (1 + 1) * 2
var d = $(c) * 2
``` 

> `var` can only be used in global and fn scopes

The `<-` operator is used for variable rewriting (usually called assignment in other languages)

```go
var a = "foo"
a <- "bar"
// <- Rewriting the variable, the value of the variable 'a' becomes bar
```

> `<-` can be used in global, fn, for scopes

## fn
fn configures a function and configures the parameters required for the function to run, such as:

```go
// t is the function alias
// time is the real function name
fn t = time {
    args = {
        "format": "YYYY-MM-DD hh:mm:ss"
    }
}
``` 

args is a built-in function configuration item, which represents the parameters passed to the function when the function is running. The fixed type of function parameters is string-to-string KVs, which corresponds to map[string]string in Go language, and the same for other languages. :warning: Note: The parameter KV received by each function is different, you need to check the specific usage of the function.

> * In the definition of `fn`, the function alias and the real function name cannot be the same
> * `fn` can only be used in the global scope

## co
co is taken from the prefix of coroutine, and is also similar to the go keyword of the Go language. The co keyword is to start running a function. For example: use co to run the print function, output Hello World!

```go
fn p = print {
    args = {
         "_" : "Hello World!" 
    }
}

co p

```
About function arguments, in the above example function p, `fn + args` is used to pass arguments to the function; in addition to using args in fn to pass in arguments, you can also pass arguments to the function directly in the co statement, e.g.:

````go
fn p = print {
}

co p {
    "_": "Hello World!"
}
````

:warning: Note: The functions do not necessarily to be defined first with fn statement, When using co to execute them (the purpose of fn is to change the default running behavior of the function through configuration), e.g.:

````go
// The print here is not the function alias defined by fn, but the real function name
co print {
    "_": "Hello World!"
}
````

About the return value of function execution, the return value of the function is the same as the arguments of the function, it is a string-to-string KVs structure, that is to say, each function will store its own return value in a map[string]string-like in the structure.

An example of getting the return value of a function:
````go
// set an out variable to receive the return value
var out
// The '->' operator represents the return of the function, so here we can think that out is a variable similar to map[string]string (actually not)
co time -> out

co print {
    // $(out.now) is the value whose key is now in the out KVs (now is a kv of the return value of the time function)
    "_": "$(out.now)"
}
````

Multiple functions can be combined in a flowl source file, so co provides the ability to execute multiple functions serially and in parallel.

```go
// serial execution
co funciton1
co function2
co function3
```

```go
// parallel execution
co {
    function1
    function2
    function3
}
```

```go
// Serial-parallel hybrid
co function1
co {
    function2
    function3
}
```

> `co` can only be used in global, for, switch scopes

## switch

`switch + case` can choose to execute `co` according to the condition. A case statement contains a conditional expression and a co statement. The following switch statement has two cases:

```go
switch { 
    case $(build) == "true" { 
        co print {
            "go build": "starting to run ..."
        }
    }
    case $(test) == "true" {
        co print {
            "go test": "starting to run ..."
        }
    }
}
```

:warning: Note: As long as the case condition in switch is true, it will be executed, which means that multiple case statements may be executed at one time, or even all of them; it does not stop when matching a case.

> `switch` can be used in global and for scopes

## event
The `event` statement is used to define an event trigger. When the trigger generates an event, it will trigger the entire flowl to be executed.

```go
event {
    co event_tick -> ev {
        "duration": "10s"
    }
    co event_cron -> ev {
        "expr": "*/5 * * * * *"
    }
}
```

In the event statement, use the co statement to start one or more event functions, which will always wait for the event to occur.

## for loop
In theory, the `for` statement in flowl, the frequency of using is not too high. In a Flow, we can use the `for` statement to control a function to be executed multiple times.

A for example with condition:
```go
var counter = 0

for $(counter) < 10 {
     // counter increments by 1
     counter <- $(counter) + 1

     // print the value of counter
     co print {
         "_": "$(counter)"
     }

     // execute function sleep, default sleep 1s
     co sleep
}
```

The for statement can also implement an infinite loop without a conditional expression:
```go
for {

}
```