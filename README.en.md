# Nestos-Assembler

### Description
`nestos-assembler` is a build environment used to build NestOS systems.

### Summary

`nestos-assembler` is a build environment that contains a series of tools that can be used to build NestOS. `nestos-assembler` implements that the process of building and testing the operating system is encapsulated in a container.

`nestos-assembler` can be simply understood as a container environment that can build nestos. This environment integrates some scripts, RPM packages and tools required to build NestOS.

### Usage

#### Build Container image
```
git clone https://gitee.com/openeuler/nestos-assembler.git
cd nestos-assembler/
docker build -f Dockerfile . -t nestos-assembler:your_tag
```
#### nosa shell script
The role of nosa shell script is to encapsulate the `nestos-assembler` call process and simplify command execution complexity. Due to differences in user usage of container runtime and building container image names, the nosa shell script is currently not provided by `nestos-assembler`. Please follow the following steps to modify and implement it in the environment where `nestos-assembler` will running:
- Write the content as shown in the following example. Please modify the container image name according to the actual situation:
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
- Save or move the script to the dir /usr/local/bin/
- Granting executable permissions:
```
sudo chmod +x /usr/local/bin/nosa
```

### Common commands
#### Basic build process
|  Command   |   Description  |
| --- | --- |
| nosa init  |  Initialize the build working directory and pull the build configuration   |
| nosa fetch  |  Update the latest build configuration and download the required rpm package for caching   |
| nosa build  |  Building a new commit of the OSTree file system, generate ociarchive file |

#### Build manage
|  Command   |   Description  |
| --- | --- |
| nosa list  |  List historical builds and built releases under the current working directory   |
| nosa clean  |  Delete all historical builds(builds/、tmp/)   |
| nosa prune  |  Delete specific build versions |
| nosa compress | Compress build artifacts |
| nosa decompress | Decompression build artifacts |
| nosa uncompress | Decompression build artifacts  |
| nosa tag | Manage build tags |

#### Build for specified platform
|  Command   |   Description  |
| --- | --- |
| nosa buildextend-qemu| Build qemu platform qcow2 format image, also can be completed by the `build` command |
| nosa buildextend-metal| Build a raw format disk image, also can be completed by the `build` command |
| nosa buildextend-metal4k| Build a native 4k mode raw format disk image, also can be completed by the `build` command |
| nosa buildextend-live| To build an ISO image with a live environment, the metal and metal4k format images must have already been built |
| nosa buildextend-openstack| Building a qcow2 image for the openstack |

#### Others
|  Command   |   Description  |
| --- | --- |
| nosa kola run | Automated functional testing of specified version builds using the Kola testing framework |
| nosa kola testiso |Automated scenario testing of building different platform artifacts for specified versions using the Kola testing framework(e.g. iso, PXE)|
| nosa kola-run | Equivalent to `kola run`, this command can eliminate interference from invalid logs during the testing process, obtain test statistics results, and extract valid logs|
| nosa push-container | Push OCI format OSTree to registry |
| nosa run | Run the NestOS qemu instance of the specified build, usually used for debugging and validation |
| nosa shell | Enter the `nestos-assembler` container image environment bash, usually used for debugging and verification |

### LICENSE

`nestos-assembler` complies with the Apache 2.0 copyright agreement.

### Notice

`nestos-assembler` is a fork of coreos-assembler(https://github.com/coreos/coreos-assembler), It will be adapted and maintained in the openEuler ecosystem, and independent evolution will be considered in the later stage.Thanks for the coreos-assembler project from Fedora coreos team.

### The main differences with coreos-assembler

Refer to [changelog-compared-with-upstream.md](./docs/changelog-compared-with-upstream.md)。
