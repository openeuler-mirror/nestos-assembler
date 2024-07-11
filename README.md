# nestos-assembler

### 介绍
nestos-assembler是用于构建NestOS系统的构建环境。

### 概述
nestos-assembler是一个构建环境，该环境包含一系列工具，可用来构建NestOS，nestos-assembler实现了在构建和测试操作系统的过程都是封装在一个容器中。

nestos-assembler可以简单理解为是一个可以构建NestOS的容器环境，该环境集成了构建NestOS所需的一些脚本、rpm包和工具。

### 使用方法

#### 容器镜像构建
```
git clone https://gitee.com/openeuler/nestos-assembler.git
cd nestos-assembler/
docker build -f Dockerfile . -t nestos-assembler:your_tag
```
#### nosa脚本
nosa 脚本的作用是封装nestos-assembler调用过程，简化命令执行复杂度。受限于用户使用容器引擎及构建容器镜像名称的不同，当前nestos-assembler暂未提供nosa脚本，请按照以下步骤在nestos-assembler运行环境中自行修改实现：
- 编写nosa脚本内容，参考示例如下，注意根据实际容器镜像名称修改：
```
#!/bin/bash

sudo docker run --rm  -it --security-opt label=disable --privileged --user=root                        \
           -v ${PWD}:/srv/ --device /dev/kvm --device /dev/fuse --network=host                                 \
           --tmpfs /tmp -v /var/tmp:/var/tmp -v /root/.ssh/:/root/.ssh/   -v /etc/pki/ca-trust/:/etc/pki/ca-trust/                                        \
           ${COREOS_ASSEMBLER_CONFIG_GIT:+-v $COREOS_ASSEMBLER_CONFIG_GIT:/srv/src/config/:ro}   \
           ${COREOS_ASSEMBLER_GIT:+-v $COREOS_ASSEMBLER_GIT/src/:/usr/lib/coreos-assembler/:ro}  \
           ${COREOS_ASSEMBLER_CONTAINER_RUNTIME_ARGS}                                            \
           ${COREOS_ASSEMBLER_CONTAINER:-nestos-assembler:your_tag} "$@"
```
- 将该脚本保存或移动至 /usr/local/bin/ 目录
- 赋予可执行权限：
```
sudo chmod +x /usr/local/bin/nosa
```

### 常用命令
#### 基本构建流程
|  命令   |   描述  |
| --- | --- |
| nosa init  |  初始化构建工作目录，拉取构建配置   |
| nosa fetch  |  更新最新构建配置，下载缓存所需rpm包   |
| nosa build  |  构建新版ostree文件系统  |

#### 构建管理
|  命令   |   描述  |
| --- | --- |
| nosa list  |  列出当前工作目录下历史构建及已构建发布件   |
| nosa clean  |  删除全部历史构建(builds、tmp)   |
| nosa prune  |  删除特定构建版本 |
| nosa compress | 压缩构建发布件 |
| nosa decompress | 解压缩构建发布件 |
| nosa uncompress | 解压缩构建发布件 |
| nosa tag | 管理构建版本标识 |

#### 特定平台发布件构建
|  命令   |   描述  |
| --- | --- |
| nosa buildextend-qemu| 构建qemu平台qcow2格式镜像，可由`build`命令添加目标同步完成 |
| nosa buildextend-metal| 构建raw格式磁盘镜像，可由`build`命令添加目标同步完成 |
| nosa buildextend-metal4k| 构建原生4k模式raw格式磁盘镜像，可由`build`命令添加目标同步完成 |
| nosa buildextend-live| 构建带有live环境的ISO镜像，必须已构建完毕metal和metal4k格式镜像 |
| nosa buildextend-openstack| 构建适用于openstack环境的qcow2镜像 |

#### 其他
|  命令   |   描述  |
| --- | --- |
| nosa kola run | 使用kola测试框架对指定版本构建进行自动化功能测试 |
| nosa kola testiso |使用kola测试框架对指定版本构建不同平台发布件进行自动化场景测试(e.g. iso, PXE)|
| nosa kola-run | 等效于`kola run`，通过此命令可排除测试过程无效日志干扰，获取测试统计结果及有效日志摘录|
| nosa push-container | 推送OCI格式ostree文件系统至容器镜像仓库 |
| nosa run | 运行指定构建版本的NestOS qemu实例，一般用于调试验证 |
| nosa runc | 以指定构建版本的NestOS 容器镜像运行命令，默认为bash，一般用于调试验证 |
| nosa shell | 进入nestos-assembler容器镜像环境bash，一般用于调试验证 |

### LICENSE

nestos-assembler 遵从 Apache 2.0 版权协议

### 声明

nestos-assembler 为 [coreos-assembler](https://github.com/coreos/coreos-assembler) 衍生版本，将在openEuler生态内适配维护，后期考虑独立演进。

感谢Fedora CoreOS团队对 coreos-assembler 的精彩付出。

### 与coreos-assembler主要差异

参见 [对比上游项目主要改动](./docs/changelog-compared-with-upstream.md)。
