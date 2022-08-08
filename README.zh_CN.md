# CoFUNC
CoFUNC 是一个基于函数编织的自动化引擎，它通过函数（function）的组合使用从而可以构建出各种能力的自动化函数流。flowl 是 CoFUNC 内嵌的一门函数编织语言，从语言层面提供函数事件、函数运行以及管理等功能。

## :rocket: FlowL
Flowl 是一门小语言，专用于函数编织； 语法非常少，也非常简单。目前已经支持函数 load，函数配置 fn，函数运行、变量定义、字符串嵌入变量、for 循环、switch 条件语句等。

### Hello World
helloworld.flowl 代码内容：
```go
// cat examples/helloworld.flowl

load "go:print"

var a = "hello world!!!"

co print {
    "_" : "$(a)"
}

```

运行代码：

```
➜ cofunc run examples/helloworld.flowl
hello world!!!
```

flowl 代码文件需要使用 `.flowl` 扩展后缀才能够被执行。

### 语法介绍
#### 注释
使用 `//` 添加代码注释。:warning: 注意，只提供独占行的注释，不能行尾注释。

#### :balloon: load
load 用于加载一个函数，例如：加载打印函数 print

```go
// go 是函数驱动，表示 print 这个函数是一段 Go 代码，需要用 go 驱动来运行
// print 是函数名
load go:print
```

所有函数在使用前，都需要先 load。

#### :balloon: fn
fn 配置一个函数，配置函数运行时需要的参数等，比如：

```go
// t 是函数别名
// time 是真实函数名
fn t = time {
    args = {
        "format": "YYYY-MM-DD hh:mm:ss"
    }
}
``` 

args 是一个内置的函数配置项，代表函数运行时传给函数的参数，函数参数固定类型为 string-to-string KVs， 对应 Go 语言就是 map[string]string，其他语言同理。:warning: 注意：每一个函数接收的参数 KV 都不一样，需要查看函数的具体用法。

#### :balloon: co
co 取自于 coroutine 的前缀，也比较类似于 Go 语言的 go 关键字。co 关键是启动运行一个函数。比如：使用 co 运行 print 函数，输出 Hello World!

```go
fn p = print {
    args = {
         "_" : "Hello World!" 
    }
}

co p
```

一个 flowl 源码文件中可以组合使用多个 function，因此 co 提供串行和并行执行多个 function 的能力。

```go
// 串行执行
co funciton1
co function2
co function3
```

```go
// 并行执行
co {
    function1
    function2
    function3
}
```

```go
// 串并行混合
co function1
co {
    function2
    function3
}
```

#### :balloon: for 循环
TODO:

#### :balloon: 变量
`var` 关键字可以定义一个变量，:warning: 注意：变量本身是没有类型的，但内置默认区分处理字符串和数字，数字变量能够进行算术运算

```go
var a = "Hello World!"
var b = $(a)
var c = (1 + 1) * 2
var d = $(c) * 2
``` 

:warning: 注意： var 可以在 global、fn、for 作用域里使用，不能在 run 里使用。

`<-` 操作符用于变量重写 （其他语言里一般叫赋值）

```go
var a = "foo"
a <- "bar
// <- 重写变量后，a 变量的值就变成了 bar
```

## :bullettrain_side: 标准函数库
- :white_check_mark: print
- :black_square_button: sleep
- :white_check_mark: command
- :white_check_mark: time
- :black_square_button: git
- :black_square_button: github
- :black_square_button: gobuild
- :black_square_button: HTTP Request
- ...

标准库的支持完全是根据我个人的日常使用工具来安排

## :bangbang: 一些重要的设计规则
TODO:

## :pushpin: TODOs
语言
* 支持条件选择语法
* 支持循环语句
* fn 作用域内支持 var 定义变量
* 支持 number 变量类型以及算术表达式
* ...

Driver
* 支持 shell driver
* 支持 Javascript driver
* 支持 Rust driver
* 支持 Docker driver
* 支持 Kubernetes driver
* ...

工具
* 函数用法
* 函数开发架手架
* cofunc-server
* repository

## 安装及配置
TODO:

## 架构设计
### 核心概念
![](docs/assets/cofunc-core-concept.png)

CoFUNC 架构设计中有 4 个核心概念，分别是 `Flow`, `Node`, `Driver` 和 `Function`

* Flow 就是用一个 `.flowl` 文件编写、定义的一条流
* Node 就是组成一条 Flow 的实体，实际执行、管理 Function 的对象
* Driver 是位于底层真正执行 Function 代码的地方，它定义了一个 Function 如何开发，如何运行，在哪里运行等等；比如：当我们需要增加 Rust 语言来开发 Function，那么就需要先实现一个 Rust 的 Driver
* Function 就是真正的函数了，它可以是一个 Go package 代码、一个二进制程序、一个 shell 脚本，或者一个 Docker 镜像等等

### flowl
![](docs/assets/flowl-parser.png)

flowl 采用词法和语法分离的实现方式，再语法分析完成得出一颗 AST 树后，再将 AST 转换成函数的运行队列，基于运行队列就可以按序执行函数

## 开发贡献
TODO:
