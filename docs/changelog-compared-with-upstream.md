# 对比上游项目主要改动

## 功能接口变化

### `nestos-assembler`不支持下列命令：
- nosa aliyun-replicate
- nosa aws-replicate
- nosa buildextend-aliyun
- nosa buildextend-aws
- nosa buildextend-azure
- nosa buildextend-azurestack
- nosa buildextend-dasd
- nosa buildextend-digitalocean
- nosa buildextend-exoscale
- nosa buildextend-gcp
- nosa buildextend-ibmcloud
- nosa buildextend-nutanix
- nosa buildextend-powervs
- nosa buildextend-vmware
- nosa buildextend-vultr
- nosa buildfetch
- nosa buildinitramfs-fast
- nosa buildupload
- nosa dev-overlay
- nosa dev-synthesize-osupdate
- nosa dev-synthesize-osupdatecontainer
- nosa generate-hashlist
- nosa koji-upload
- nosa oc-adm-release
- nosa remote-prune
- nosa sign
- nosa test-coreos-installer
- nosa upload-oscontainer

### `nestos-assembler`添加下列命令：
- nosa kola-run

## 代码变化

#### 新增构建变种NestOS，并设为默认
#### 默认构建base镜像及软件源修改为openEuler
#### 新增iSula容器引擎测试用例
#### push-container命令增强
支持添加“--insecure”参数，忽略目标镜像仓库的SSL/TLS验证


## 目录结构变化

### 移除与上游CICD相关代码目录及构建内容：
- ci/
- gangplank/
- docs/gangplank/

### 修改和新增部分文档说明
