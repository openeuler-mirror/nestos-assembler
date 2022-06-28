# nestos-assembler

#### Description
nestos-installer is a build environment used to build NestOS systems.

#### Summary

nestos-assembler is a build environment that contains a series of tools that can be used to build NestOS. nestos-assembler implements that the process of building and testing the operating system is encapsulated in a container.

nestos-assembler can be simply understood as a container environment that can build nestos. This environment integrates some scripts, RPM packages and tools required to build NestOS.

#### Common commands

|  name   |   Description  |
| --- | --- |
|   nosa clean  |  Delete all build artifacts   |
|   nosa fetch  |  Fetch and import the latest packages   |
|   nosa build  |   Generate qemu artifacts for the given platforms  |
| nosa buildextend-metal | Generate metal artifacts for the given platforms |
| nosa buildextend-metal4k | Generate metal4k artifacts for the given platforms |
| nosa buildextend-live | Generate the Live ISO |

#### LICENSE

nestos-assembler complies with the Apache 2.0 copyright agreement.

#### Notice

nestos-assembler is a fork of coreos-assembler(https://github.com/coreos/coreos-assembler), It will be adapted and maintained in the openEuler ecosystem, and independent evolution will be considered in the later stage.Thanks for the coreos-assembler project from Fedora coreos team.
