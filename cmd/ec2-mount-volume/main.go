// Copyright (C) 2019 The ec2-mount-volume Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Command ec2-mount-volume prepares EBS volumes, exposed as NVMe block devices, for immediate use. It mounts each volume at the location specified in a resource tag.
//
// It provides tag-based mapping to work around unpredictable device names. For example, an instance that mounts at boot a root volume and one data volume may encounter device names "/dev/nvme0n1" and "/dev/nvme1n1" after one boot, but reversed names after the next.
//
// Usage:
//
//   ec2-mount-volume --help
//
// Display the mount plan for 2 expected EBS volumes (dry run):
//
//   ec2-mount-volume --device-num 2
//
// Mount 2 expected EBS volumes:
//
//   ec2-mount-volume --device-num 2 --force
//
// Same as above but read mount points from EBS volume tags named "mount-point" instead of the default:
//
//   ec2-mount-volume --device-num 2 --force --tag mount-point
//
// Same as above but wait 30 seconds instead of the default:
//
//   ec2-mount-volume --device-num 2 --force --tag mount-point --timeout 30
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	tp_ec2 "github.com/codeactual/ec2-mount-volume/internal/third_party/github.com/aws/ec2"
	"github.com/jpillora/backoff"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/codeactual/ec2-mount-volume/internal/cage/aws/v1/ec2/metadata"
	"github.com/codeactual/ec2-mount-volume/internal/cage/cli/handler"
	handler_cobra "github.com/codeactual/ec2-mount-volume/internal/cage/cli/handler/cobra"
	cage_exec "github.com/codeactual/ec2-mount-volume/internal/cage/os/exec"
	cage_reflect "github.com/codeactual/ec2-mount-volume/internal/cage/reflect"
)

const (
	defaultTagName = "Mount"
	symlinkPrefix  = "nvme-Amazon_Elastic_Block_Store_vol"
	volumeIdPrefix = "vol-"

	defaultPartSuffix = "p1"
	defaultMountOpt   = "defaults"
	defaultFsType     = "ext4"
	defaultTimeout    = 60
)

func main() {
	err := handler_cobra.NewHandler(&Handler{}).Execute()
	if err != nil {
		panic(errors.WithStack(err))
	}
}

// Handler defines the sub-command flags and logic.
type Handler struct {
	handler.IO

	DeviceNum  int    `yaml:"Expected number of volumes"`
	Force      bool   `yaml:"Disable the default dry-run mode"`
	FsType     string `yaml:"Filesystem type"`
	MountOpt   string `yaml:"'mount' option list"`
	PartSuffix string `yaml:"Suffix appended to device paths to create partition paths"`
	Tag        string `yaml:"Name of the EBS volume resource tag specifying mount points"`
	Timeout    uint   `usage:"Number of seconds to wait for all volumes to be mounted before cancellation"`
}

// Init defines the command, its environment variable prefix, etc.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) Init() handler_cobra.Init {
	return handler_cobra.Init{
		Cmd: &cobra.Command{
			Use:   "ec2-mount-volume",
			Short: "Mount /dev/disk/by-id/* volumes mapped by their tags",
		},
		EnvPrefix: "EC2_MOUNT_VOLUME",
	}
}

// BindFlags binds the flags to Handler fields.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) BindFlags(cmd *cobra.Command) []string {
	cmd.Flags().IntVarP(&h.DeviceNum, "device-num", "", 0, cage_reflect.GetFieldTag(*h, "DeviceNum", "usage"))
	cmd.Flags().BoolVarP(&h.Force, "force", "", false, cage_reflect.GetFieldTag(*h, "Force", "usage"))
	cmd.Flags().StringVarP(&h.FsType, "fs-type", "", defaultFsType, cage_reflect.GetFieldTag(*h, "FsType", "usage"))
	cmd.Flags().StringVarP(&h.MountOpt, "mount-opt", "", defaultMountOpt, cage_reflect.GetFieldTag(*h, "MountOpt", "usage"))
	cmd.Flags().StringVarP(&h.PartSuffix, "part-suffix", "", defaultPartSuffix, cage_reflect.GetFieldTag(*h, "PartSuffix", "usage"))
	cmd.Flags().StringVarP(&h.Tag, "tag", "", defaultTagName, cage_reflect.GetFieldTag(*h, "Tag", "usage"))
	cmd.Flags().UintVarP(&h.Timeout, "timeout", "", defaultTimeout, cage_reflect.GetFieldTag(*h, "Timeout", "usage"))
	return []string{"device-num"}
}

// Run performs the sub-command logic.
//
// It implements cli/handler/cobra.Handler.
func (h *Handler) Run(ctx context.Context, args []string) {
	dryrun := !h.Force

	maxTries := 10.0
	b := &backoff.Backoff{Jitter: true}

	var volumes []*ec2.Volume

	// Example error to tolerate with backoff:
	// https://ec2.us-west-2.amazonaws.com/: dial tcp: lookup ec2.us-west-2.amazonaws.com on [::1]:53: read udp [::1]:38986->[::1]:53: read: connection refused
	for {
		var err error

		volumes, err = metadata.Volumes()
		if err != nil {
			if b.Attempt() == maxTries {
				panic(errors.WithStack(errors.Wrapf(err, fmt.Sprintf("failed to get volumes metadata after %d tries", int(maxTries)))))
			}

			d := b.Duration()
			fmt.Fprintf(h.Err(), "failed to get volumes metadata [%s], trying again after %s", err.Error(), d.String())
			time.Sleep(d)

			continue
		}

		break
	}

	var fsckArgs [][]string
	var mountArgs [][]string

	ctx, cancel := context.WithTimeout(ctx, time.Duration(h.Timeout)*time.Second)
	defer cancel()

	for _, volume := range volumes {
		var mount string
		for _, tag := range volume.Tags {
			if *tag.Key == h.Tag && *tag.Value != "" {
				mount = *tag.Value
				break
			}
		}

		if mount == "" {
			continue
		}

		device, err := tp_ec2.FindNvmeVolume(symlinkPrefix + strings.TrimLeft(*volume.VolumeId, volumeIdPrefix))
		if err != nil {
			panic(errors.WithStack(err))
		}

		// -M: error if already mounted; -y: attempt to repair issues; -V: verbose
		fsckArgs = append(fsckArgs, []string{"fsck", "-M", "-y", "-V", device + h.PartSuffix})

		mountArgs = append(mountArgs, []string{
			"mount",
			"-o", h.MountOpt,
			"-t", h.FsType,
			device + h.PartSuffix,
			mount,
		})
	}

	actualDeviceNum := len(fsckArgs)
	if h.DeviceNum != actualDeviceNum {
		fmt.Fprintf(os.Stderr, "Canceled. Expected %d devices to mount but detected %d.\n", h.DeviceNum, actualDeviceNum)
		os.Exit(1)
	}

	for _, args := range fsckArgs {
		if dryrun {
			fmt.Println(strings.Join(args, " "))
			continue
		}
		mustExecCmd(ctx, args)
	}

	for _, args := range mountArgs {
		if dryrun {
			fmt.Println(strings.Join(args, " "))
			continue
		}
		mustExecCmd(ctx, args)
	}

	if dryrun {
		fmt.Println("Dry run complete. Run with --force to execute the commands.")
	}
}

func mustExecCmd(ctx context.Context, args []string) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	_, err := cage_exec.CommonExecutor{}.Standard(ctx, os.Stdout, os.Stderr, nil, cmd)
	if err != nil {
		panic(errors.WithStack(err))
	}
}

var _ handler_cobra.Handler = (*Handler)(nil)
