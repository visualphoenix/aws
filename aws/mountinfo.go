package aws

import (
	"fmt"
	"github.com/visualphoenix/disk-go/lsblk"
	"github.com/visualphoenix/disk-go/fs"
	"github.com/visualphoenix/disk-go/lvm"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// MountInfo contants the union of lvm and aws volume info
type MountInfo struct {
	Mountpoint string
	FilesystemType string
	BlockDevice string
	BlockDeviceType string
	PhysicalDevices []VolumeInfo
}

func getVolume(volumes []VolumeInfo, disk string) VolumeInfo {
	for _, v := range volumes {
		if v.Device == disk {
			return v
		}
	}
	return VolumeInfo{}
}

// GetMountInfoFromVolumes returns a list of MountInfo from lsblk output and a list of attached aws volumes
func GetMountInfoFromVolumes(l lsblk.Lsblk, volumes []VolumeInfo) []MountInfo {
	var result []MountInfo;
	mountpointToDisks := make(map[string][]VolumeInfo)
	for _, d := range l.Disks {
		if d.Disk.Mountpoint != "" {
			if _, ok := mountpointToDisks[d.Disk.Mountpoint]; !ok {
				m := MountInfo {
					Mountpoint: d.Disk.Mountpoint,
					FilesystemType: d.Disk.Fstype,
					BlockDevice: d.Disk.Device,
					BlockDeviceType: d.Disk.Dtype,
				}
				result = append(result, m)
			}
			mountpointToDisks[d.Disk.Mountpoint] =  append(mountpointToDisks[d.Disk.Mountpoint],getVolume(volumes,d.Disk.Device))
		}
		for _, p := range d.Parts {
			if p.Mountpoint != "" {
				if _, ok := mountpointToDisks[p.Mountpoint]; !ok {
					m := MountInfo {
						Mountpoint: p.Mountpoint,
						FilesystemType: p.Fstype,
						BlockDevice: p.Device,
						BlockDeviceType: p.Dtype,
					}
					result = append(result, m)
				}
				mountpointToDisks[p.Mountpoint] = append(mountpointToDisks[p.Mountpoint],getVolume(volumes,d.Disk.Device))
			}
		}
	}
	for i := range result {
		disks := mountpointToDisks[result[i].Mountpoint]
		result[i].PhysicalDevices = disks
	}
	return result
}

// GetMountInfoFrom returns a list of MountInfo for a given aws instanceID
func GetMountInfoFrom(service *ec2.EC2, instanceID string) ([]MountInfo, error) {
	l, err := lsblk.GetLsblkInfo()
	if err != nil {
		return []MountInfo{}, fmt.Errorf("parse error: %s", err)
	}
	volumes, err := GetAttachedVolumes(service, instanceID)
	if err != nil {
		return []MountInfo{}, err
	}
	mi := GetMountInfoFromVolumes(l, volumes)
	return mi, nil
}

// Suspend writes to the device/partition given the type of the mount
func (mi MountInfo) Suspend() {
	switch mi.BlockDeviceType {
	case "disk":
		fallthrough
	case "part":
		fs.Freeze(mi.Mountpoint)
	case "lvm":
		lvm.Suspend(mi.BlockDevice)
	default:
	}
}

// Resume writes to the device/partition given the type of the mount
func (mi MountInfo) Resume() {
	switch mi.BlockDeviceType {
	case "disk":
		fallthrough
	case "part":
		fs.Freeze(mi.Mountpoint)
	case "lvm":
		lvm.Suspend(mi.BlockDevice)
	default:
	}
}
