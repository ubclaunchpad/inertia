package provision

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/awslabs/aws-sdk-go/service/iam"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
)

// EC2Provisioner creates Amazon EC2 instances
type EC2Provisioner struct {
	user    string
	session *session.Session
	client  *ec2.EC2
}

// NewEC2Provisioner creates a client to interact with Amazon EC2 using the
// given credentials
func NewEC2Provisioner(id, key string) (*EC2Provisioner, error) {
	prov := &EC2Provisioner{}
	return prov, prov.init(credentials.NewStaticCredentials(id, key, ""))
}

// NewEC2ProvisionerFromEnv creates a client to interact with Amazon EC2 using
// credentials from environment
func NewEC2ProvisionerFromEnv() (*EC2Provisioner, error) {
	prov := &EC2Provisioner{}
	return prov, prov.init(credentials.NewEnvCredentials())
}

// GetUser returns the user attached to given credentials
func (p *EC2Provisioner) GetUser() string { return p.user }

// ListRegions lists available regions to create an instance in
func (p *EC2Provisioner) ListRegions() ([]string, error) {
	// Set an arbitrary region, since the API requires this
	p.client.Config.WithRegion("us-east-1")

	// Get list of available regions
	regions, err := p.client.DescribeRegions(nil)
	if err != nil {
		return nil, err
	}
	regionList := []string{}
	for _, r := range regions.Regions {
		regionList = append(regionList, *r.RegionName)
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
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("name"),
				Values: []*string{aws.String("amzn*")},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Sort by date
	sort.Slice(output.Images, func(i, j int) bool {
		iCreated := common.ParseDate(*output.Images[i].CreationDate)
		if iCreated == nil {
			return false
		}
		jCreated := common.ParseDate(*output.Images[j].CreationDate)
		if jCreated == nil {
			return true
		}
		return iCreated.After(*jCreated)
	})

	images := []string{}
	for _, image := range output.Images {
		if len(images) == 10 {
			break
		}
		// Ignore nameless images
		if image.Name != nil {
			images = append(images, fmt.Sprintf("%s (%s)", *image.ImageId, *image.Description))
		}
	}
	return images, nil
}

// CreateInstance creates an EC2 instance with given properties
func (p *EC2Provisioner) CreateInstance(name, imageID, instanceType, region string) (*cfg.RemoteVPS, error) {
	// Set requested region
	p.client.Config.WithRegion(region)

	// Generate authentication
	keyName := fmt.Sprintf("%s_%s_inertia_key_%d", name, p.user, time.Now().UnixNano())
	fmt.Printf("Generating key pair %s...", keyName)
	keyResp, err := p.client.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: aws.String(keyName),
	})
	if err != nil {
		return nil, err
	}

	// Save key
	keyPath := filepath.Join(os.Getenv("HOME"), ".ssh", *keyResp.KeyName)
	fmt.Printf("Saving key to %s...", keyPath)
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
	println(runResp.GoString())

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
	println("Checking status of requested instance...")
	attempts := 0
	var instanceStatus *ec2.DescribeInstancesOutput
	for true {
		attempts++
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
			println("Instance status: " + *result.Reservations[0].Instances[0].State.Name)
			time.Sleep(3 * time.Second)
			continue
		} else {
			// Code 16 means we can continue!
			println("Instance is running!")
			instanceStatus = result
			break
		}
	}

	// Return remote configuration
	return &cfg.RemoteVPS{
		Name:    name,
		IP:      *instanceStatus.Reservations[0].Instances[0].PublicDnsName,
		User:    p.user,
		PEM:     keyPath,
		SSHPort: "22",
	}, nil
}

func (p *EC2Provisioner) init(creds *credentials.Credentials) error {
	// Set up configuration
	p.session = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	config := &aws.Config{Credentials: creds}

	// Attempt to access credentials
	identityClient := iam.New(p.session, config)
	user, err := identityClient.GetUser(nil)
	if err != nil {
		return err
	}
	p.user = *user.User.UserName

	// Set up EC2 client
	p.client = ec2.New(p.session, config)
	// Workaround for a strange bug where client instantiates with "https://ec2..amazonaws.com"
	p.client.Endpoint = "https://ec2.amazonaws.com"
	return nil
}
