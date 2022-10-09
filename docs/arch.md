# Architecture Design

<div align=center><img width="50%" height="50%" src="assets/arch.png"/></div>

## Runtime Core Concepts

<div align=center><img width="50%" height="50%" src="assets/cofunc-core-concept.png"/></div>

There are 4 core concepts in cofx architecture design when it's running, namely `Flow`, `Node`, `Driver` and `Function`

* `Flow` is a process that's defined through a `.flowl` file
* `Node` is the entity that makes up a Flow, the node entity executes and manages a Function
* `Driver` is the place where the function code is actually executed. It defines how a function is developed, how to run, where to run, etc. For example, when we need to add Rust language to develop functions, then we need to implement a Rust driver first
* `Function` is the real function, it maybe a Go package code, a binary program, a shell script, or a Docker image, etc.

## flowL

<div align=center><img width="70%" height="70%" src="assets/flowl-parser.png"/></div>

flowl adopts the implementation method of lexical and grammar separation. After the grammar, it will output an AST tree, the AST is converted into a run queue of functions. Based on the run queue, functions can be executed in order.