package aws
// Copyright 2018 Raymond Barbiero. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	_aws "github.com/aws/aws-sdk-go/aws"
)

// SnapshotInfo configures the properties of the snapshot we want to take
type SnapshotInfo struct {
	VolumeID string
	Description string
	Tags TagMap
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
				Tags:         s.Tags.ToEC2Tags(),
			},
		},
		DryRun: _aws.Bool(s.DryRun),
	})
}
