package provision

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
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

// ListImageOptions lists available Amazon images for your given region
func (p *EC2Provisioner) ListImageOptions(region string) ([]string, error) {
	// Set requested region
	p.client.Config.WithRegion(region)

	// Query for images from the Amazon
	output, err := p.client.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{aws.String("amazon")},
		Filters: []*ec2.Filter{
			{
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

// EC2CreateInstanceOptions defines parameters with which to create an EC2 instance
type EC2CreateInstanceOptions struct {
	Name        string
	ProjectName string
	Ports       []int64
	DaemonPort  int64

	ImageID      string
	InstanceType string
	Region       string

	User string
}

// CreateInstance creates an EC2 instance with given properties
func (p *EC2Provisioner) CreateInstance(opts EC2CreateInstanceOptions) (*cfg.RemoteVPS, error) {
	// Set requested region
	p.client.Config.WithRegion(opts.Region)

	// Set user if given
	if opts.User != "" {
		p.user = opts.User
	}

	// Generate authentication
	keyName := fmt.Sprintf("%s_%s_inertia_key_%d", opts.Name, p.user, time.Now().UnixNano())
	fmt.Printf("Generating key pair %s...\n", keyName)
	keyResp, err := p.client.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: aws.String(keyName),
	})
	if err != nil {
		return nil, err
	}

	// Save key
	keyPath := filepath.Join(os.Getenv("HOME"), ".ssh", *keyResp.KeyName)
	fmt.Printf("Saving key to %s...\n", keyPath)
	err = local.SaveKey(*keyResp.KeyMaterial, keyPath)
	if err != nil {
		return nil, err
	}

	// Create security group for network configuration
	group, err := p.client.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName: aws.String(
			fmt.Sprintf("%s-%s-%d", opts.ProjectName, opts.Name, time.Now().UnixNano()),
		),
		Description: aws.String(
			fmt.Sprintf("Rules for project %s on %s", opts.ProjectName, opts.Name),
		),
	})
	if err != nil {
		return nil, err
	}

	// Set rules for ports
	err = p.exposePorts(*group.GroupId, opts.DaemonPort, opts.Ports)
	if err != nil {
		return nil, err
	}

	// Start up instance
	runResp, err := p.client.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(opts.ImageID),
		InstanceType: aws.String(opts.InstanceType),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),

		// Security options
		KeyName:          keyResp.KeyName,
		SecurityGroupIds: []*string{group.GroupId},
	})
	if err != nil {
		return nil, err
	}

	// Loop until intance is running
	println("Checking status of requested instance...")
	attempts := 0
	var instanceStatus *ec2.DescribeInstancesOutput
	for {
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

	// Set tags
	_, err = p.client.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{runResp.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(opts.Name),
			},
			{
				Key:   aws.String("Purpose"),
				Value: aws.String("Inertia Continuous Deployment"),
			},
		},
	})

	// Poll for SSH port to open
	println("Waiting for port 22 to open...")
	for {
		time.Sleep(3 * time.Second)
		println("Checking port...")
		conn, err := net.Dial("tcp", *instanceStatus.Reservations[0].Instances[0].PublicDnsName+":22")
		if err == nil {
			println("Connection established!")
			conn.Close()
			break
		}
	}

	// Generate webhook secret
	webhookSecret, err := common.GenerateRandomString()
	if err != nil {
		println(err.Error())
		webhookSecret = "interia"
	}

	// Return remote configuration
	return &cfg.RemoteVPS{
		Name:    opts.Name,
		IP:      *instanceStatus.Reservations[0].Instances[0].PublicDnsName,
		User:    p.user,
		PEM:     keyPath,
		SSHPort: "22",
		Daemon: &cfg.DaemonConfig{
			Port:          strconv.FormatInt(opts.DaemonPort, 10),
			WebHookSecret: webhookSecret,
		},
	}, nil
}

// exposePorts updates the security rules of given security group to expose
// given ports
func (p *EC2Provisioner) exposePorts(securityGroupID string, daemonPort int64, ports []int64) error {
	// Create Inertia rules
	portRules := []*ec2.IpPermission{{
		FromPort:   aws.Int64(int64(22)),
		ToPort:     aws.Int64(int64(22)),
		IpProtocol: aws.String("tcp"),
		IpRanges:   []*ec2.IpRange{{CidrIp: aws.String("0.0.0.0/0"), Description: aws.String("Inertia SSH port")}},
		Ipv6Ranges: []*ec2.Ipv6Range{{CidrIpv6: aws.String("::/0"), Description: aws.String("Inertia SSH port")}},
	}, {
		FromPort:   aws.Int64(daemonPort),
		ToPort:     aws.Int64(daemonPort),
		IpProtocol: aws.String("tcp"),
		IpRanges:   []*ec2.IpRange{{CidrIp: aws.String("0.0.0.0/0"), Description: aws.String("Inertia daemon port")}},
		Ipv6Ranges: []*ec2.Ipv6Range{{CidrIpv6: aws.String("::/0"), Description: aws.String("Inertia daemon port")}},
	}}

	// Generate rules for user project
	for _, port := range ports {
		portRules = append(portRules, &ec2.IpPermission{
			FromPort:   aws.Int64(port),
			ToPort:     aws.Int64(port),
			IpProtocol: aws.String("tcp"), // todo: allow config
			IpRanges:   []*ec2.IpRange{{CidrIp: aws.String("0.0.0.0/0")}},
			Ipv6Ranges: []*ec2.Ipv6Range{{CidrIpv6: aws.String("::/0")}},
		})
	}

	// Set rules
	_, err := p.client.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:       aws.String(securityGroupID),
		IpPermissions: portRules,
	})
	return err
}

func (p *EC2Provisioner) init(creds *credentials.Credentials) error {
	// Set default user
	p.user = "ec2-user"

	// Set up configuration
	p.session = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Set up EC2 client
	p.client = ec2.New(p.session, &aws.Config{Credentials: creds})
	// Workaround for a strange bug where client instantiates with "https://ec2..amazonaws.com"
	p.client.Endpoint = "https://ec2.amazonaws.com"
	return nil
}
