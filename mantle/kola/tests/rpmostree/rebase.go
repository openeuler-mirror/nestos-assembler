package rpmostree

import (
	"github.com/coreos/mantle/kola/cluster"
	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/kola/tests/util"
)

func init() {
	register.RegisterTest(&register.Test{
		Run:         rpmOstreeRebase,
		ClusterSize: 1,
		Name:        "rpmostree.rebase",
		FailFast:    true,
		Tags:        []string{"rpm-ostree", "upgrade"},
		Flags:       []register.Flag{register.RequiresInternetAccess},
		// remove this testcase for iso,becase ro mount 'error: Remounting /sysroot read-write: Permission denied'
		// ExcludePlatforms: []string{"qemu-iso"},
	})
}

func rpmOstreeRebase(c cluster.TestCluster) {
	m := c.Machines()[0]
	arch := c.MustSSH(m, "uname -m")
	var newBranch string = "ostree-unverified-registry:hub.oepkgs.net/nestos/nestos:22.03-LTS-SP3.20240110.0-" + string(arch)

	originalStatus, err := util.GetRpmOstreeStatusJSON(c, m)
	if err != nil {
		c.Fatal(err)
	}

	if len(originalStatus.Deployments) < 1 {
		c.Fatalf(`Unexpected results from "rpm-ostree status"; received: %v`, originalStatus)
	}

	c.Run("ostree upgrade", func(c cluster.TestCluster) {
		// use "rpm-ostree rebase" to get to the "new" commit
		_ = c.MustSSH(m, "sudo systemctl start docker.service && sudo rpm-ostree rebase --experimental "+newBranch+" --bypass-driver")
		// get latest rpm-ostree status output to check validity
		postUpgradeStatus, err := util.GetRpmOstreeStatusJSON(c, m)
		if err != nil {
			c.Fatal(err)
		}

		// should have an additional deployment
		if len(postUpgradeStatus.Deployments) != len(originalStatus.Deployments)+1 {
			c.Fatalf("Expected %d deployments; found %d deployments", len(originalStatus.Deployments)+1, len(postUpgradeStatus.Deployments))
		}
		// reboot into new deployment
		rebootErr := m.Reboot()
		if rebootErr != nil {
			c.Fatalf("Failed to reboot machine: %v", rebootErr)
		}

		// get latest rpm-ostree status output
		postRebootStatus, err := util.GetRpmOstreeStatusJSON(c, m)
		if err != nil {
			c.Fatal(err)
		}

		// should have 2 deployments, the previously booted deployment and the test deployment due to rpm-ostree pruning
		if len(postRebootStatus.Deployments) != 2 {
			c.Fatalf("Expected %d deployments; found %d deployment", 2, len(postRebootStatus.Deployments))
		}

		// new deployment should be booted
		if !postRebootStatus.Deployments[0].Booted {
			c.Fatalf("New deployment is not reporting as booted")
		}
	})
}
