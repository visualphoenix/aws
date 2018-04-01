package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	_aws "github.com/aws/aws-sdk-go/aws"
)

// SnapshotInfo configures the properties of the snapshot we want to take
type SnapshotInfo struct {
	VolumeID string
	Description string
	Tags map[string]string
	DryRun bool
}

// CreateSnapshot wraps the aws CreateSnapshot to be more golang
func CreateSnapshot(e *ec2.EC2, s SnapshotInfo) (*ec2.Snapshot, error) {
	return e.CreateSnapshot(&ec2.CreateSnapshotInput{
		VolumeId: _aws.String(s.VolumeID),
		Description: _aws.String(s.Description),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: _aws.String(ec2.ResourceTypeSnapshot),
				Tags:         getTags(s.Tags),
			},
		},
		DryRun: _aws.Bool(s.DryRun),
	})
}

func getTags(tags map[string]string) []*ec2.Tag {
	var result []*ec2.Tag
	for k, v := range tags {
		result = append(result, &ec2.Tag{Key: _aws.String(k), Value: _aws.String(v)})
	}
	return result
}
