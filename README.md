![](./docs/assets/logo.png)

[[中文](./README.zh_CN.md)]

CoFx is an automation engine that uses low-code programming to build personal automation workflows, so that turn boring suff into low code. The CoFx framework engine consists of two parts, the programming language and the standard function library.

![](./docs/assets/demo.gif)

## Features
* Built-in function fabric language flowL
* Commonly used function standard library
* CLI tool that conforms to programmer's habits
* Use low-code to develop workflow through function fabric
* Built-in out-of-the-box workflow by default
* Support extended development of functions
* Support event trigger workflow
* ...

## Default Built-in Workflow 
* [github-3way-sync](docs/github_3way_sync.md)：Automatically synchronize the local, origin and upstream branches of the Github project
* [github-auto-pr](docs/github_auto_pr.md)：Automatically push the local branch to origin, then automatically create a pull request and open the pull request details page through a browser
* [go-auto-build](docs/go_auto_build.md)：Automatically build a go project based on 'go module', support automatic detection of multiple modules, automatic build
* ...

Install cofx and use the `cofx list` command to view all default built-in workflows.

## Standard Library Functions

| Function Name        | Explain                                                      |
| :------------------- | :----------------------------------------------------------- |
| command              | Run a command or script                                      |
| print                | Print to stdout                                              |
| time                 | Read the current time and return multiple time value related variables |
| event/event_cron     | Timing event trigger based on cron syntax format             |
| event/event_tick     | Timing event trigger at fixed interval X seconds             |
| git/git_add_upstream | Use the 'git remote add' command to configure upstream       |
| git/git_check_merge  | Use 'git merge-base/merge-tree' command to check two branches for conflict |
| git/git_fetch        | Use the 'git fetch' command to update the local repository   |
| git/git_local_info   | Read common basic information of local git repository        |
| git/git_pull         | Update the local repository with the 'git pull' command      |
| git/git_push         | Sync local branch to remote using 'git push' command         |
| git/git_rebase       | Merge branches using 'git rebase' command                    |
| github/gh_create_pr  | Create a pull request to upstream                            |
| go/go_build          | Analyze the go project of 'go mod' and automatically build each module |
| go/go_generate       | Wraps the 'go generate' command                              |
| go/go_test           | Wraps the 'go test' unit testing command                     |
| http/http_get        | Send a HTTP GET request                                      |
| http/http_post       | Send a HTTP POST request                                     |
| ...                  |                                                              |

Install cofx, use the `cofx std` command to view all the functions of the standard library; use the `cofx std <function name>` to view the specific usage of the function's parameters and return values.

## flowL - A small language
Flowl is a small language that be used to `function fabric`; The syntax is very minimal and simple. Currently, it supports function load, function configuration, function operation, variable definition and operation, embedded variable into string, for loop, switch conditional statement, etc.

#### Hello World
helloworld.flowl code content:
```go
// cat examples/helloworld.flowl

load "go:print"

var a = "hello world!!!"

co print {
    "_" : "$(a)"
}
```

Run the code:

![](./docs/assets/hello.png)

The flowL source file needs to use the `.flowl` extension to be executed.

[FlowL Grammar Introduction](docs/flowl_guide.md)

## Installation
#### MacOS 

```
brew tap cofxlabs/tap
brew install cofx
```

#### Universal Installation
Download the appropriate latest version from Release and execute the following command to install:

```
tar zxvf cofx-<your-os-arch>.tar.gz
cd <your-os-arch>
sudo ./install.sh
```

## Develop&Contribition
* [Archtecture Design](docs/arch.md)
