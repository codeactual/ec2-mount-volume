package ec2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindNvmeVolume looks for the nvme volume with the specified name
// It follows the symlink (if it exists) and returns the absolute path to the device
//
// Origin:
//   https://github.com/kubernetes/kubernetes/blob/f4472b1a92877ed4b1576e7e44496b0de7a8efe2/pkg/volume/aws_ebs/aws_util.go#L227
//   Apache 2.0: https://github.com/kubernetes/kubernetes/blob/f4472b1a92877ed4b1576e7e44496b0de7a8efe2/LICENSE
//
// Changes:
// - Remove k8s log lines
func FindNvmeVolume(findName string) (device string, err error) {
	const (
		defaultTagName = "Mount"
		symlinkPrefix  = "nvme-Amazon_Elastic_Block_Store_vol"
		volumeIdPrefix = "vol-"
	)

	p := filepath.Join("/dev/disk/by-id/", findName)
	stat, err := os.Lstat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("error getting stat of %q: %v", p, err)
	}

	if stat.Mode()&os.ModeSymlink != os.ModeSymlink {
		return "", nil
	}

	// Find the target, resolving to an absolute path
	// For example, /dev/disk/by-id/nvme-Amazon_Elastic_Block_Store_vol0fab1d5e3f72a5e23 -> ../../nvme2n1
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		return "", fmt.Errorf("error reading target of symlink %q: %v", p, err)
	}

	if !strings.HasPrefix(resolved, "/dev") {
		return "", fmt.Errorf("resolved symlink for %q was unexpected: %q", p, resolved)
	}

	return resolved, nil
}
