package aws

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
    // LogOff states that no logging should be performed by the SDK. This is the
    // default state of the SDK, and should be use to disable all logging.
    LogOff = aws.LogOff

    // LogDebug state that debug output should be logged by the SDK. This should
    // be used to inspect request made and responses received.
    LogDebug = aws.LogDebug
)
const (
    // LogDebugWithSigning states that the SDK should log request signing and
    // presigning events. This should be used to log the signing details of
    // requests for debugging. Will also enable LogDebug.
    LogDebugWithSigning = aws.LogDebugWithSigning

    // LogDebugWithHTTPBody states the SDK should log HTTP request and response
    // HTTP bodys in addition to the headers and path. This should be used to
    // see the body content of requests and responses made while using the SDK
    // Will also enable LogDebug.
    LogDebugWithHTTPBody = aws.LogDebugWithHTTPBody

    // LogDebugWithRequestRetries states the SDK should log when service requests will
    // be retried. This should be used to log when you want to log when service
    // requests are being retried. Will also enable LogDebug.
    LogDebugWithRequestRetries = aws.LogDebugWithRequestRetries

    // LogDebugWithRequestErrors states the SDK should log when service requests fail
    // to build, send, validate, or unmarshal.
    LogDebugWithRequestErrors = aws.LogDebugWithRequestRetries
)

// Options contain the options for aws plugin
type Options struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Retries         int
	LogLevel		aws.LogLevelType
}

// MetadataKey is the identifier of a metadata entry
type MetadataKey string

const (
	// MetadataInstanceID - Instance ID
	MetadataInstanceID = MetadataKey("http://169.254.169.254/latest/meta-data/instance-id")

	// MetadataAvailabilityZone - Availability Zone
	MetadataAvailabilityZone = MetadataKey("http://169.254.169.254/latest/meta-data/placement/availability-zone")
)

// GetMetadata returns the value of a metadata key
func GetMetadata(key MetadataKey) (string, error) {
	resp, err := http.Get(string(key))
	if err != nil {
		return "", err
	}
	buff, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}
	return string(buff), nil
}

// MetadataAvailable returns true if the host can reach the aws metadata ip address
func MetadataAvailable() bool {
	_, err := net.DialTimeout("tcp", "169.254.169.254:80", 1*time.Second)
	return err == nil
}

// GetRegion returns the AWS region of this instance
func GetRegion() (string, error) {
	az, err := GetMetadata(MetadataAvailabilityZone)
	if err != nil {
		return "", err
	}
	return az[0 : len(az)-1], nil
}

var customLogger aws.Logger
func init() {
	SetLogger(defaultLogger{logger: log.New(os.Stderr, "", log.LstdFlags)})
}

func SetLogger(l aws.Logger) {
	customLogger = l
}

type defaultLogger struct {
	logger* log.Logger
}

func (l defaultLogger) Log(args ...interface{}) {
	l.logger.Println(args...)
}

// GetLogger gets a logger that can be used with the AWS SDK.
func GetLogger() aws.Logger {
	return customLogger
}

// GetService returns a new EC2 service
func GetService(config *session.Session) (*ec2.EC2) {
	return ec2.New(config)
}

// GetInstance returns an ec2.Instance pointer for a given instanceID
func GetInstance(service *ec2.EC2, instanceID string) (*ec2.Instance, error) {
	input := ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceID),
		},
	}

	result, err := service.DescribeInstances(&input)
	if err != nil {
		return nil, err
	}

	if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("Could not find an instance with ID: %s", instanceID)
	}

	return result.Reservations[0].Instances[0], nil
}

// GetSession returns an AWS session pointer given an options config struct
func GetSession(o Options) *session.Session {
	providers := []credentials.Provider{
		&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.Must(session.NewSession()))},
		&credentials.EnvProvider{},
		&credentials.SharedCredentialsProvider{},
	}

	if (len(o.AccessKeyID) > 0 && len(o.SecretAccessKey) > 0) || len(o.SessionToken) > 0 {
		staticCreds := credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     o.AccessKeyID,
				SecretAccessKey: o.SecretAccessKey,
				SessionToken:    o.SessionToken,
			},
		}
		providers = append(providers, &staticCreds)
	}

	creds := credentials.NewChainCredentials(providers)
	region, _ := GetRegion()
	config := aws.NewConfig().WithRegion(region).WithLogger(GetLogger()).
				WithMaxRetries(11).
				WithCredentialsChainVerboseErrors(true).WithCredentials(creds).
				WithLogLevel(aws.LogLevelType(o.LogLevel))
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            *config,
	}
	session := session.Must(session.NewSessionWithOptions(opts))

	return session
}
