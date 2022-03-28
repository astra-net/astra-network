# Introduction

This document introduces the Astra's package release using standard packaging system, RPM and Deb packages.

Standard packaging system has many benefits, like extensive tooling, documentation, portability, and complete design to handle different situation.

# Package Content

The RPM/Deb packages will install the following files/binary in your system.

- /usr/sbin/astra
- /usr/sbin/astra-setup.sh
- /usr/sbin/astra-rclone.sh
- /etc/astra/astra.conf
- /etc/astra/rclone.conf
- /etc/systemd/system/astra.service
- /etc/sysctl.d/99-astra.conf

The package will create `astra` group and `astra` user on your system.
The astra process will be run as `astra` user.
The default blockchain DBs are stored in `/home/astra/astra_db_?` directory.
The configuration of astra process is in `/etc/astra/astra.conf`.

# Package Manager

Please take sometime to learn about the package managers used on Fedora/Debian based distributions.
There are many other package managers can be used to manage rpm/deb packages like [Apt]<https://en.wikipedia.org/wiki/APT_(software)>,
or [Yum]<https://www.redhat.com/sysadmin/how-manage-packages>

# Setup customized repo

You just need to do the setup of astra repo once on a new host.
**TODO**: the repo in this document are for development/testing purpose only.

Official production repo will be different.

## RPM Package

RPM is for Redhat/Fedora based Linux distributions, such as Amazon Linux and CentOS.

```bash
# do the following once to add the astra development repo
curl -LsSf http://haochen-astra-pub.s3.amazonaws.com/pub/yum/astra-dev.repo | sudo tee -a /etc/yum.repos.d/astra-dev.repo
sudo rpm --import https://raw.githubusercontent.com/astra-net/astra-network-open/master/astra-release/astra-pub.key
```

## Deb Package

Deb is supported on Debian based Linux distributions, such as Ubuntu, MX Linux.

```bash
# do the following once to add the astra development repo
curl -LsSf https://raw.githubusercontent.com/astra-net/astra-network-open/master/astra-release/astra-pub.key | sudo apt-key add
echo "deb http://haochen-astra-pub.s3.amazonaws.com/pub/repo bionic main" | sudo tee -a /etc/apt/sources.list

```

# Test cases

## installation

```
# debian/ubuntu
sudo apt-get update
sudo apt-get install astra

# fedora/amazon linux
sudo yum install astra
```

## configure/start

```
# dpkg-reconfigure astra (TODO)
sudo systemctl start astra
```

## uninstall

```
# debian/ubuntu
sudo apt-get remove astra

# fedora/amazon linux
sudo yum remove astra
```

## upgrade

```bash
# debian/ubuntu
sudo apt-get update
sudo apt-get upgrade

# fedora/amazon linux
sudo yum update --refresh
```

## reinstall

```bash
remove and install
```

# Rclone

## install latest rclone

```bash
# debian/ubuntu
curl -LO https://downloads.rclone.org/v1.52.3/rclone-v1.52.3-linux-amd64.deb
sudo dpkg -i rclone-v1.52.3-linux-amd64.deb

# fedora/amazon linux
curl -LO https://downloads.rclone.org/v1.52.3/rclone-v1.52.3-linux-amd64.rpm
sudo rpm -ivh rclone-v1.52.3-linux-amd64.rpm
```

## do rclone

```bash
# validator runs on shard1
sudo -u astra astra-rclone.sh /home/astra 0
sudo -u astra astra-rclone.sh /home/astra 1

# explorer node
sudo -u astra astra-rclone.sh -a /home/astra 0
```

# Setup explorer (non-validating) node

To setup an explorer node (non-validating) node, please run the `astra-setup.sh` at first.

```bash
sudo /usr/sbin/astra-setup.sh -t explorer -s 0
```

to setup the node as an explorer node w/o blskey setup.

# Setup new validator

Please copy your blskey to `/home/astra/.astra/blskeys` directory, and start the node.
The default configuration is for validators on mainnet. No need to run `astra-setup.sh` script.

# Start/stop node

- `systemctl start astra` to start node
- `systemctl stop astra` to stop node
- `systemctl status astra` to check status of node

# Change node configuration

The node configuration file is in `/etc/astra/astra.conf`. Please edit the file as you needed.

```bash
sudo vim /etc/astra/astra.conf
```

# Support

Please open new github issues in https://github.com/astra-net/astra-network/issues.
