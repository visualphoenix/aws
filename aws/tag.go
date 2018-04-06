package aws

import (
	"fmt"
	"log"
	"strings"
	"github.com/aws/aws-sdk-go/service/ec2"
	_aws "github.com/aws/aws-sdk-go/aws"
)
// TagInfo describes the tags for an aws resource
type TagInfo struct {
	Resource string
	Tags TagMap
	DryRun bool
}

// CreateTags tags an aws resource
func CreateTags(service *ec2.EC2, ti TagInfo) (*ec2.CreateTagsOutput, error) {
		p := &ec2.CreateTagsInput{
			Resources: []*string{
				_aws.String(ti.Resource),
			},
			Tags: ti.Tags.ToEC2Tags(),
			DryRun: _aws.Bool(ti.DryRun),
		}

		result, err := service.CreateTags(p)
		if err != nil {
			return nil, err
		}
		return result, nil
}

// TagMap is a map of aws tag key/values
type TagMap map[string]string

// String strinifies the output of a TagMap
func (m *TagMap) String() string {
	var result string
	for key, val := range *m {
		result = result + fmt.Sprintf("%s=\"%s\"\n", key, val)
	}
	return strings.TrimSuffix(result, "\n")
}

// Set takes a key=value string and assigns the value to the
// key in the TagMap
func (m *TagMap) Set(keyvalue string) error {
	i := strings.Index(keyvalue, "=")
	if i <= -1 {
		log.Fatal(keyvalue, "does not contain =")
	} else {
		key := keyvalue[:i]
		val := keyvalue[i+1:]
		(*m)[key] = val
	}
	return nil
}

// MergeTags merges the tags from b into a's TagMap
func MergeTags(a TagMap, b TagMap) TagMap {
	result := a
	for key, val := range b {
		result[key] = val
	}
	return result
}

// ToEC2Tags converts a TagMap to an array of ec2.Tag
func (m *TagMap) ToEC2Tags() []*ec2.Tag {
	var result []*ec2.Tag
	for k, v := range *m {
		result = append(result, &ec2.Tag{Key: _aws.String(k), Value: _aws.String(v)})
	}
	return result
}

// NewTagMap creates a TagMap from an array of ec2.Tag
func NewTagMap(tags []*ec2.Tag) TagMap {
	result := make(TagMap)
	result["Name"] = ""
	for _,tag := range tags {
		result[*tag.Key] = *tag.Value
	}
	return result
}
