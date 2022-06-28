# nestos-assembler

#### 介绍
nestos-assembler是用于构建NestOS系统的构建环境.

#### 概述
nestos-assembler是一个构建环境，该环境包含一系列工具，可用来构建NestOS，nestos-assembler实现了在构建和测试操作系统的过程都是封装在一个容器中。

nestos-assembler可以简单理解为是一个可以构建NestOS的容器环境，该环境集成了构建NestOS所需的一些脚本、rpm包和工具。

#### 常用命令
|  name   |   Description  |
| --- | --- |
|   nosa clean  |  删除历史构建(builds、tmp)   |
|   nosa fetch  |  下载所需rpm包   |
|   nosa build  |   通过下载的rpm包构建ostree和qemu镜像  |
| nosa buildextend-metal | 构建metal镜像 |
| nosa buildextend-metal4k | 构建metal4k镜像 |
| nosa buildextend-live | 构建iso镜像 |

#### LICENSE

nestos-assembler 遵从 Apache 2.0 版权协议

#### 说明

nestos-assembler 基于 coreos-assembler(https://github.com/coreos/coreos-assembler) 分叉，将在openEuler生态内适配维护，后期考虑独立演进。
