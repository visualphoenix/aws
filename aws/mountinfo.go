package aws
// Copyright 2018 Raymond Barbiero. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"fmt"
	"github.com/visualphoenix/disk-go/fs"
	"github.com/visualphoenix/disk-go/lvm"
	"github.com/visualphoenix/disk-go/lsblk"
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

type blockDeviceDiskMap map[string][]VolumeInfo

func getDiskVolumeInfo(volumes []VolumeInfo, disk string) VolumeInfo {
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
	blockDeviceToDisks := make(blockDeviceDiskMap)
	for _, d := range l.Disks {
		disk := getDiskVolumeInfo(volumes,d.Disk.Device)
		if d.Disk.Fstype != "LVM2_member" && d.Disk.Fstype != "" {
			if _, ok := blockDeviceToDisks[d.Disk.Device]; !ok {
				m := MountInfo {
					Mountpoint: d.Disk.Mountpoint,
					FilesystemType: d.Disk.Fstype,
					BlockDevice: d.Disk.Device,
					BlockDeviceType: d.Disk.Dtype,
				}
				result = append(result, m)
			}
			blockDeviceToDisks[d.Disk.Device] =  append(blockDeviceToDisks[d.Disk.Device], disk)
		}
		for _, p := range d.Parts {
			if p.Fstype != "LVM2_member" {
				if _, ok := blockDeviceToDisks[p.Device]; !ok {
					m := MountInfo {
						Mountpoint: p.Mountpoint,
						FilesystemType: p.Fstype,
						BlockDevice: p.Device,
						BlockDeviceType: p.Dtype,
					}
					result = append(result, m)
				}
				blockDeviceToDisks[p.Device] = append(blockDeviceToDisks[p.Device], disk)
			}
		}
	}
	for i := range result {
		disks := blockDeviceToDisks[result[i].BlockDevice]
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
func (mi MountInfo) Suspend() error {
	var err error
	if mi.BlockDeviceType == "disk" || mi.BlockDeviceType == "part" {
		if mi.Mountpoint != "" {
			err = fs.Freeze(mi.Mountpoint)
		}
	} else if mi.BlockDeviceType == "lvm" {
		err = lvm.Suspend(mi.BlockDevice)
	}
	return err
}

// Resume writes to the device/partition given the type of the mount
func (mi MountInfo) Resume() error {
	var err error
	if mi.BlockDeviceType == "disk" || mi.BlockDeviceType == "part" {
		if mi.Mountpoint != "" {
			err = fs.Unfreeze(mi.Mountpoint)
		}
	} else if mi.BlockDeviceType == "lvm" {
		err = lvm.Resume(mi.BlockDevice)
	}
	return err
}
