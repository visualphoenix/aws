package aws

import (
	"strings"
	"errors"
	"path/filepath"
	"io/ioutil"
	"log"
	"github.com/aws/aws-sdk-go/service/ec2"
	_aws "github.com/aws/aws-sdk-go/aws"
)

// VolumeInfo describes an aws volume in native golang
type VolumeInfo struct {
	Device string
	State string
	InstanceID string
	VolumeID string
	Tags map[string]string
}

// GetAttachedVolumes queries an aws instance for the attached volumes and returns a list of VolumeInfo
func GetAttachedVolumes(e *ec2.EC2, instanceID string) ([]VolumeInfo, error) {
	volumes, err := e.DescribeVolumes(&ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name: _aws.String("attachment.instance-id"),
				Values: []*string{
					_aws.String(instanceID),
				},
			},
		},
	})

	var results []VolumeInfo
	if err != nil {
		return results, err
	}

	for _, volume := range volumes.Volumes {
		if len(volume.Attachments) == 1 {
			tags := make(map[string]string)
			tags["Name"] = ""
			for _,tag := range volume.Tags {
				tags[*tag.Key] = *tag.Value
			}
			v := VolumeInfo {
				Device: getDevice(*volume.Attachments[0].Device),
				State: *volume.Attachments[0].State,
				InstanceID: *volume.Attachments[0].InstanceId,
				VolumeID: *volume.Attachments[0].VolumeId,
				Tags: tags,
			}
			results = append(results, v)
		}
	}
	return results, nil
}

func localDevicePrefix() (string, error) {
	available := []string{"sd", "xvd"}

	f, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		log.Fatal(err)
	}
	for _, d := range f {
		base := filepath.Base(d.Name())
		for _, prefix := range available {
			if strings.HasPrefix(base, prefix) {
				return prefix, nil
			}
		}
	}

	return "", errors.New("device prefix could not be detected")
}

func getDevice(device string) string {
	prefix, _ := localDevicePrefix()
	d := filepath.Base(device)
	if d == "sda1" && prefix == "xvd" {
		d = "xvda"
	}
	if strings.HasPrefix(d, "sd") {
		return prefix + d[2:]
	}

	if strings.HasPrefix(d, "xvd") {
		return prefix + d[3:]
	}

	return device
}
