# 对比上游项目主要改动

## 支持架构变化
当前NestOS仅支持x86_64与aarch64两种架构

## 功能接口变化

### `nestos-assembler`不支持下列命令：
- nosa aliyun-replicate
- nosa aws-replicate
- nosa buildextend-aliyun
- nosa buildextend-applehv
- nosa buildextend-aws
- nosa buildextend-azure
- nosa buildextend-azurestack
- nosa buildextend-dasd
- nosa buildextend-digitalocean
- nosa buildextend-exoscale
- nosa buildextend-gcp
- nosa buildextend-hyperv
- nosa buildextend-ibmcloud
- nosa buildextend-kubevirt
- nosa buildextend-nutanix
- nosa buildextend-powervs
- nosa buildextend-virtualbox
- nosa buildextend-vmware
- nosa buildextend-vultr
- nosa buildfetch
- nosa buildinitramfs-fast
- nosa buildupload
- nosa dev-overlay
- nosa dev-synthesize-osupdate
- nosa dev-synthesize-osupdatecontainer
- nosa koji-upload
- nosa oc-adm-release
- nosa powervs-replicate
- nosa remote-prune
- nosa sign
- nosa update-variant
- nosa upload-oscontainer

### `nestos-assembler`添加下列命令：
- nosa kola-run

## 代码变化

#### 新增构建变种NestOS，并设为默认
#### 默认构建base镜像及软件源修改为openEuler
#### 新增iSula容器引擎测试用例
#### push-container命令增强
- 支持添加`--no-tls-verify`参数，忽略目标镜像仓库的SSL/TLS验证
- 支持添加`-d`或`--dest-repository`参数，通过参数指定推送目标容器镜像仓库
- 支持添加`--transport`参数，指定推送容器镜像传输方式，可通过此方式推送到本地容器引擎
#### 修复以tcg方式构建live镜像失败问题
#### 屏蔽不支持的kola选项
#### 暂不支持以osbuild方式构建磁盘镜像


## 目录结构变化

### 移除与上游CICD相关代码目录及构建内容：
- ci/
- continuous/

### 修改和新增部分文档说明
