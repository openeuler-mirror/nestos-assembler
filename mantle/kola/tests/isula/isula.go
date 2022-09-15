package isula

import (
	"fmt"
	"strings"
	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/kola/cluster"
	"github.com/coreos/mantle/platform"
	tutil "github.com/coreos/mantle/kola/tests/util"
)

func init(){
	register.RegisterTest(&register.Test{
		Run:         isulaBaseTest,
		ClusterSize: 1,
		Name:        `isula.base`,
	})
}

func isulaBaseTest(c cluster.TestCluster) {
	c.Run("info", isulaInfo)
	c.Run("resources", isulaResources)
}

// Test: Verify basic isula info information
func isulaInfo(c cluster.TestCluster) {
	m := c.Machines()[0]
	_, err := c.SSH(m, `sudo systemctl start isulad.service`)
	if err != nil {
		c.Fatal(err)
	}
	err = getIsulaInfo(c, m)
	if err != nil {
		c.Fatal(err)
	}
}

// Returns the result of isula info as a simplifiedIsulaInfo
func getIsulaInfo(c cluster.TestCluster, m platform.Machine) (error) {

	result, err := c.SSH(m, `sudo isula info`)
	if err != nil {
		return fmt.Errorf("Could not get info: %v", err)
	}

	for _, line := range strings.Split(string(result), "\n") {
		if strings.Contains(line, "Storage Driver") {
			if strings.Contains(line, "overlay") {
				continue
			} else{
				c.Errorf("Unexpected Storage Driver")
			}
		}
		if strings.Contains(line, "Operating System") {
			if strings.Contains(line, "NestOS") {
				continue
			} else{
				c.Errorf("Unexpected Operating System")
			}
		}
	}
	return nil
}

// Test: Run isula with various options
func isulaResources(c cluster.TestCluster) {
	m := c.Machines()[0]

	_, err := c.SSH(m, `sudo systemctl start isula-build.service`)
	if err != nil {
		c.Fatal(err)
	}

	tutil.GenIsulaScratchContainer(c, m, "echo", []string{"echo"})

	isulaFmt := "sudo isula run --net=none --rm %s echo1 echo 1"

	pCmd := func(arg string) string {
		return fmt.Sprintf(isulaFmt, arg)
	}

	for _, isulaCmd := range []string{
		// must set memory when setting memory-swap
		// See https://github.com/opencontainers/runc/issues/1980 for
		// why we use 128m for memory
		pCmd("--memory=128m --memory-swap=128m"),
		pCmd("--memory-reservation=10m"),
		pCmd("--cpu-shares=100"),
		pCmd("--cpu-period=1000"),
		pCmd("--cpuset-cpus=0"),
		pCmd("--cpuset-mems=0"),
		pCmd("--cpu-quota=1000"),
		pCmd("--memory=128m"),
		pCmd("--shm-size=1m"),
	} {
		cmd := isulaCmd
		output, err := c.SSH(m, cmd)
		if err != nil {
			c.Fatalf("Failed to run %q: output: %q status: %q", cmd, output, err)
		}
	}
}