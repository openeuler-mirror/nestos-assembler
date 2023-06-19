#!/usr/bin/env bash
set -euo pipefail

# Keep this script idempotent for local development rebuild use cases:
# any consecutive runs should produce the same result.

arch=$(uname -m)

if [ $# -gt 1 ]; then
  echo Usage: "build.sh [CMD]"
  echo "Supported commands:"
  echo "    configure_user"
  echo "    configure_yum_repos"
  echo "    install_rpms"
  echo "    make_and_makeinstall"
  exit 1
fi

set -x
srcdir=$(pwd)

configure_yum_repos() {
    local version_id
    version_id=$(. /etc/os-release && echo ${VERSION_ID})
    # Add continuous tag for latest build tools and mark as required so we
    # can depend on those latest tools being available in all container
    # builds.
    #echo -e "[${version_id}-nestos]\nenabled=1\nmetadata_expire=1m\nbaseurl=http://10.1.110.88/nestos-assembler/\ngpgcheck=0\nskip_if_unavailable=False\n" > /etc/yum.repos.d/nestos.repo
    rm -rf /etc/yum.repos.d/*
    # openeuler 22.03-LTS
    echo -e "[${version_id}]\nenabled=1\nmetadata_expire=1m\nbaseurl=http://119.3.219.20:82/openEuler:/22.03:/LTS/standard_$arch/\ngpgcheck=0\npriority=1\nskip_if_unavailable=False\n" > /etc/yum.repos.d/nestos-LTS.repo
    echo -e "[${version_id}-Next]\nenabled=1\nmetadata_expire=1m\nbaseurl=http://119.3.219.20:82/openEuler:/22.03:/LTS:/Epol/standard_$arch/\ngpgcheck=0\npriority=1\nskip_if_unavailable=False\n" > /etc/yum.repos.d/nestos-EPOL.repo
    
    echo -e "[${version_id}-SP1]\nenabled=1\nmetadata_expire=1m\nbaseurl=http://119.3.219.20:82/openEuler:/22.03:/LTS:/SP1/standard_$arch/\ngpgcheck=0\npriority=2\nskip_if_unavailable=False\n" > /etc/yum.repos.d/nestos-SP1.repo
    echo -e "[${version_id}-SP1-epol]\nenabled=1\nmetadata_expire=1m\nbaseurl=http://119.3.219.20:82/openEuler:/22.03:/LTS:/SP1:/Epol/standard_$arch/\ngpgcheck=0\npriority=2\nskip_if_unavailable=False\n" > /etc/yum.repos.d/nestos-sp1-epol.repo
}

install_rpms() {
    local builddeps
    local frozendeps

    # freeze kernel due to https://github.com/coreos/nestos-assembler/issues/2707
    #frozendeps=$(echo kernel-5.10.0-60.41.0.73.oe2203)

    yum install -y qemu-img qemu-block-iscsi qemu-block-curl qemu-hw-usb-host qemu-system-x86_64 qemu liburing-devel glib2-devel
    arch=$(uname -m)
    case $arch in
    "x86_64")  rpm -iUh qemu-*.x86_64.rpm libslirp-*.x86_64.rpm;;
    "aarch64")  rpm -iUh qemu-*.aarch64.rpm libslirp-*.aarch64.rpm;;
    *)         fatal "Architecture ${arch} not supported"
    esac

    yum install -y libsolv rpm-devel grubby initscripts iptables nftables python3-setuptools linux-firmware bubblewrap json-c ostree json-glib polkit-libs ostree-devel dnf-plugins-core container-selinux oci-runtime
    arch=$(uname -m)
    case $arch in
    "x86_64")  rpm -iUh libsolv-0.7.22-1.x86_64.rpm libsolv-devel-0.7.22-1.x86_64.rpm kernel-5.10.0-60.41.0.73.oe2203.x86_64.rpm kernel-headers-5.10.0-60.41.0.73.oe2203.x86_64.rpm buildah-1.26.1-1.x86_64.rpm butane-0.14.0-1.oe2203.x86_64.rpm dumb-init-1.2.5-4.oe2203.x86_64.rpm python3-semver-2.10.2-2.oe2203.noarch.rpm containers-common-1-1.oe2203.noarch.rpm netavark-1.0.2-1.x86_64.rpm rpm-ostree-2022.8-3.oe2203.x86_64.rpm rpm-ostree-devel-2022.8-3.oe2203.x86_64.rpm supermin-5.3.2-1.x86_64.rpm;;
    "aarch64")  rpm -iUh libsolv-0.7.22-1.aarch64.rpm libsolv-devel-0.7.22-1.aarch64.rpm kernel-5.10.0-118.0.0.64.oe2203.aarch64.rpm kernel-headers-5.10.0-118.0.0.64.oe2203.aarch64.rpm buildah-1.26.1-1.oe2203.aarch64.rpm butane-0.14.0-2.oe2203.aarch64.rpm dumb-init-1.2.5-1.oe2203.aarch64.rpm python3-semver-2.10.2-2.oe2203.noarch.rpm containers-common-1-1.oe2203.noarch.rpm netavark-1.0.2-1.oe2203.aarch64.rpm rpm-ostree-2022.8-3.oe2203.aarch64.rpm rpm-ostree-devel-2022.8-3.oe2203.aarch64.rpm supermin-5.3.2-1.oe2203.aarch64.rpm;;
    *)         fatal "Architecture ${arch} not supported"
    esac
    
    # First, a general update; this is best practice.  We also hit an issue recently
    # where qemu implicitly depended on an updated libusbx but didn't have a versioned
    # requires https://bugzilla.redhat.com/show_bug.cgi?id=1625641
    #yum -y distro-sync

    # xargs is part of findutils, which may not be installed
    yum -y install /usr/bin/xargs

    # These are only used to build things in here.  Today
    # we ship these in the container too to make it easier
    # to use the container as a development environment for itself.
    # Down the line we may strip these out, or have a separate
    # development version.
    builddeps=$(grep -v '^#' "${srcdir}"/src/build-deps.txt)

    # Process our base dependencies + build dependencies and install
    #(echo "${builddeps}" && echo "${frozendeps}" && "${srcdir}"/src/print-dependencies.sh) | xargs yum -y install
    (echo "${builddeps}" && "${srcdir}"/src/print-dependencies.sh) | xargs yum -y install

    # Add fast-tracked packages here.  We don't want to wait on bodhi for rpm-ostree
    # as we want to enable fast iteration there.
    #yum --enablerepo=updates-testing upgrade rpm-ostree

    # Allow Kerberos Auth to work from a keytab. The keyring is not
    # available in a Container.
    sed -e "s/^.*default_ccache_name/#    default_ccache_name/g" -i /etc/krb5.conf

    # Open up permissions on /boot/efi files so we can copy them
    # for our ISO installer image, skip if not present
    if [ -e /boot/efi ]; then
        chmod -R a+rX /boot/efi
    fi
    # Similarly for kernel data and SELinux policy, which we want to inject into supermin
    chmod -R a+rX /usr/lib/modules /usr/share/selinux/targeted
    # Further cleanup
    yum clean all
}

make_and_makeinstall() {
    make && make install
    rm -rf /root/.cache/go-build
}

configure_user(){
    # modify Time Zone
    rm -f /etc/localtime
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
    # /dev/kvm might be bound in, but will have the gid from the host, and not all distros
    # a+rw permissions on /dev/kvm. create groups for all the common kvm gids and then add
    # builder to them.
    # systemd defaults to 0666 but other packages like qemu sometimes override this with 0660.
    # Adding the user to the kvm group should always work.

    # fedora uses gid 36 for kvm
    getent group kvm78  || groupadd -g 78 -o -r kvm78   # arch, gentoo
    getent group kvm124 || groupadd -g 124 -o -r kvm124 # debian
    getent group kvm232 || groupadd -g 232 -o -r kvm232 # ubuntu

    # We want to run what builds we can as an unprivileged user;
    # running as non-root is much better for the libvirt stack in particular
    # for the cases where we have --privileged in the container run for other reasons.
    # At some point we may make this the default.
    getent passwd builder || useradd builder --uid 1000 -G wheel,kvm,kvm78,kvm124,kvm232
    echo '%wheel ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/wheel-nopasswd
    # Contents of /etc/sudoers.d need not to be world writable
    chmod 600 /etc/sudoers.d/wheel-nopasswd
}

write_archive_info() {
    cp -f ./packages /usr/lib64/guestfs/supermin.d/packages
    #head -n -1 /etc/bashrc > ./bashrc
    
    #mv -f ./bashrc /etc/bashrc
    head -n -1 /etc/bashrc
    echo "TMOUT=32400" >> /etc/bashrc
    #source /etc/bashrc
    # shellcheck source=src/cmdlib.sh
    . "${srcdir}/src/cmdlib.sh"
    mkdir -p /cosa /lib/coreos-assembler
    touch -f /lib/coreos-assembler/.clean
    prepare_git_artifacts "${srcdir}" /cosa/coreos-assembler-git.tar.gz /cosa/coreos-assembler-git.json
}

if [ $# -ne 0 ]; then
  # Run the function specified by the calling script
  ${1}
else
  # Otherwise, just run all the steps
  configure_yum_repos
  install_rpms
  write_archive_info
  make_and_makeinstall
  configure_user
fi
