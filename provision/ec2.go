package provision

import (
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/local"
)

// EC2Provisioner creates Amazon EC2 instances
type EC2Provisioner struct {
	client *ec2.EC2
}

// NewEC2Provisioner creates a client to interact with Amazon EC2 using the
// given credentials
func NewEC2Provisioner(id, key string) *EC2Provisioner {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := ec2.New(sess, &aws.Config{
		Credentials: credentials.NewStaticCredentials(id, key, ""),
	})
	// Workaround for a strange bug where client instantiates with "https://ec2..amazonaws.com"
	client.Endpoint = "https://ec2.amazonaws.com"
	return &EC2Provisioner{client: client}
}

// NewEC2ProvisionerFromEnv creates a client to interact with Amazon EC2 using
// credentials from environment
func NewEC2ProvisionerFromEnv() *EC2Provisioner {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	client := ec2.New(sess, &aws.Config{
		Credentials: credentials.NewEnvCredentials(),
	})
	// Workaround for a strange bug where client instantiates with "https://ec2..amazonaws.com"
	client.Endpoint = "https://ec2.amazonaws.com"
	return &EC2Provisioner{client: client}
}

// ListRegions lists available regions to create an instance in
func (p *EC2Provisioner) ListRegions() ([]string, error) {
	regions, err := p.client.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}

	regionList := []string{}
	for _, r := range regions.Regions {
		regionList = append(regionList, r.GoString())
	}
	return regionList, nil
}

// ListImageOptions lists available Amazon images for your given region
func (p *EC2Provisioner) ListImageOptions(region string) ([]string, error) {
	// Set requested region
	p.client.Config.WithRegion(region)

	// Query for images from the Amazon
	output, err := p.client.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{aws.String("amazon")},
	})
	if err != nil {
		return nil, err
	}

	// Return relevant list
	images := []string{}
	for _, image := range output.Images {
		// todo: improve return structure
		images = append(images, image.GoString())
	}
	return images, nil
}

// CreateInstance creates an EC2 instance with given properties
func (p *EC2Provisioner) CreateInstance(name, imageID, instanceType, region string) (*cfg.RemoteVPS, error) {
	// Set requested region
	p.client.Config.WithRegion(region)

	// Generate authentication
	keyResp, err := p.client.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: aws.String(name + "_inertia_key"),
	})
	if err != nil {
		return nil, err
	}

	// Save key
	keyPath := filepath.Join(os.Getenv("HOME"), ".ssh", *keyResp.KeyName)
	err = local.SaveKey(*keyResp.KeyMaterial, keyPath)
	if err != nil {
		return nil, err
	}

	// Start up instance
	runResp, err := p.client.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(imageID),
		InstanceType: aws.String(instanceType),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      keyResp.KeyName,
	})
	if err != nil {
		return nil, err
	}

	// Set some instance tags for convenience
	_, err = p.client.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{runResp.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
			{
				Key:   aws.String("Purpose"),
				Value: aws.String("Inertia Continuous Deployment"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Loop until intance is running
	attempts := 0
	for true {
		attempts++
		println("Checking status of requested instance...")
		// Request instance status
		result, err := p.client.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{runResp.Instances[0].InstanceId},
		})
		if err != nil {
			return nil, err
		}
		if len(result.Reservations) == 0 || len(result.Reservations[0].Instances) == 0 {
			// A reservation corresponds to a command to start instances
			time.Sleep(3 * time.Second)
			continue
		} else if *result.Reservations[0].Instances[0].State.Code != 16 {
			// Code 16 indicates instance is running
			println("Instance status: " + result.Reservations[0].Instances[0].State.GoString())
			time.Sleep(3 * time.Second)
			continue
		} else {
			// Code 16 means we can continue!
			println("Instance is running!")
			break
		}
	}

	// Get branch for remote setup
	branch, err := local.GetRepoCurrentBranch()
	if err != nil {
		println("failed to set branch - defaulting to 'master'")
		branch = "master"
	}

	// Return remote configuration
	return &cfg.RemoteVPS{
		Name:    name,
		IP:      *runResp.Instances[0].PublicIpAddress,
		User:    "ec2-user", // default user
		PEM:     keyPath,
		Branch:  branch,
		SSHPort: "22",
	}, nil
}
