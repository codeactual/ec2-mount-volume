# ec2-mount-volume [![GoDoc](https://godoc.org/github.com/codeactual/ec2-mount-volume?status.svg)](https://godoc.org/github.com/codeactual/ec2-mount-volume) [![Go Report Card](https://goreportcard.com/badge/github.com/codeactual/ec2-mount-volume)](https://goreportcard.com/report/github.com/codeactual/ec2-mount-volume) [![Build Status](https://travis-ci.org/codeactual/ec2-mount-volume.png)](https://travis-ci.org/codeactual/ec2-mount-volume)

ec2-mount-volume prepares EBS volumes, exposed as NVMe block devices, for immediate use. It mounts each volume at the location specified in a resource tag.

## Use Case

It provides tag-based mapping to work around unpredictable device names. For example, an instance that mounts at boot a root volume and one data volume may encounter device names "/dev/nvme0n1" and "/dev/nvme1n1" after one boot, but reversed names after the next.

It was made after encountering this issue while migrating from an M4 to M5 instance type.

# Usage

> To install: `go get -v github.com/codeactual/ec2-mount-volume/cmd/ec2-mount-volume`

## Examples

> Usage:

```bash
ec2-mount-volume --help
```

> Display the mount plan for 2 expected EBS volumes (dry run):

```bash
ec2-mount-volume --device-num 2
```

> Mount 2 expected EBS volumes:

```bash
ec2-mount-volume --device-num 2 --force
```

> Same as above but read mount points from EBS volume tags named "mount-point" instead of the default:

```bash
ec2-mount-volume --device-num 2 --force --tag mount-point
```

> Same as above but wait 30 seconds instead of the default:

```bash
ec2-mount-volume --device-num 2 --force --tag mount-point --timeout 30
```

# Setup

The program requires access to the EC2's metadata service and the EC2 instance role requires `ec2:DescribeVolumes` access.

# License

[Mozilla Public License Version 2.0](https://www.mozilla.org/en-US/MPL/2.0/) ([About](https://www.mozilla.org/en-US/MPL/), [FAQ](https://www.mozilla.org/en-US/MPL/2.0/FAQ/))

*(Exported from a private monorepo with [transplant](https://github.com/codeactual/transplant).)*
