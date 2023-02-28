# github-auto-pr 介绍

<img src="assets/auto-pr.png" style="zoom:67%;" />

## 场景

参与 github 开源项目的时候，我们总是会不停的写代码，然后提交 pull request。cofx github-auto-pr workflow 的目标就是本地写完代码，提交 commit 后，就可以自动化的完成最终 pull request 的创建，整个过程只需要一条命令即可完成。

## 用法

进入本地一个 clone 的 github 项目仓库内，checkout 一个开发分支，完成代码开发后，同时完成代码的 commit。再 commit 之后，就可以直接执行如下命令来完成 pull request 的自动创建。

```shell
cofx run github-auto-pr -e GITHUB_TOKEN=<YOUR-TOKEN>
```

执行 github-auto-pr workflow 后，首先会将本地 branch 通过 git push 同步到 origin 远端仓库，然后再通过 github  api 创建一个 pull request 给 upstream 远端仓库。

注意：

* github-auto-pr workflow 需要依赖 github token 做身份验证（后续，cofx 可能会提供 token 、password 管理，优化每次执行都需要反复输入 token，password 的场景）。
* 默认，pull request 是创建到 upstream 仓库的 main 分支，如果需要调整为其他分支，需要修改 github-auto-pr.flowl。 

示范：

<img src="assets/auto-pr-demo.png" style="zoom:67%;" />

## FlowL 程序

github auto pr 的 flowl 程序位于：

[https://github.com/skoowoo/cofx/blob/main/examples/github-auto-pr.flowl](https://github.com/skoowoo/cofx/blob/main/examples/github-auto-pr.flowl)

