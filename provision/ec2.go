package provision

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
	"github.com/ubclaunchpad/inertia/common"
	"github.com/ubclaunchpad/inertia/local"
)

const (
	// Code returned by AWS when EC2 instance is successfully created
	codeEC2InstanceStarted = 16
)

// EC2Provisioner creates Amazon EC2 instances
type EC2Provisioner struct {
	out     io.Writer
	user    string
	session *session.Session
	client  *ec2.EC2
}

// NewEC2Provisioner creates a client to interact with Amazon EC2 using the
// given credentials
func NewEC2Provisioner(user, keyID, key string, out ...io.Writer) (*EC2Provisioner, error) {
	prov := &EC2Provisioner{}
	return prov, prov.init(user, credentials.NewStaticCredentials(keyID, key, ""), out)
}

// NewEC2ProvisionerFromEnv creates a client to interact with Amazon EC2 using
// credentials from environment
func NewEC2ProvisionerFromEnv(user string, out ...io.Writer) (*EC2Provisioner, error) {
	prov := &EC2Provisioner{}
	return prov, prov.init(user, credentials.NewEnvCredentials(), out)
}

// NewEC2ProvisionerFromProfile creates a client to interact with Amazon EC2 using
// credentials for user (optional) from given profile file
func NewEC2ProvisionerFromProfile(user, profile, path string, out ...io.Writer) (*EC2Provisioner, error) {
	prov := &EC2Provisioner{}
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return prov, prov.init(user, credentials.NewSharedCredentials(path, profile), out)
}

// GetUser returns the user attached to given credentials
func (p *EC2Provisioner) GetUser() string { return p.user }

// ListImageOptions lists available Amazon images for your given region
func (p *EC2Provisioner) ListImageOptions(region string) ([]string, error) {
	// Set requested region
	p.WithRegion(region)

	// Query for easily supported images
	output, err := p.client.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{aws.String("amazon")},
		Filters: []*ec2.Filter{
			{
				// Only display Amazon for ease of use
				Name:   aws.String("name"),
				Values: []*string{aws.String("amzn*")},
			},
			{
				// Docker needs machine to run properly
				Name:   aws.String("image-type"),
				Values: []*string{aws.String("machine")},
			},
			{
				// No funny business
				Name:   aws.String("architecture"),
				Values: []*string{aws.String("x86_64")},
			},
			{
				// Most standard instances only support EBS
				Name:   aws.String("root-device-type"),
				Values: []*string{aws.String("ebs")},
			},
			{
				// Paravirtual images don't work - see #500
				Name:   aws.String("virtualization-type"),
				Values: []*string{aws.String("hvm")},
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

	// Format image names for printing
	images := []string{}
	for _, image := range output.Images {
		if len(images) == 10 {
			break
		}
		// Ignore nameless images
		if image.Name != nil {
			// Rudimentary filter to remove ECS images - see https://github.com/ubclaunchpad/inertia/issues/633
			if strings.Contains(*image.ImageId, "ECS") || strings.Contains(*image.Description, "ECS") {
				continue
			}

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
}

// CreateInstance creates an EC2 instance with given properties
func (p *EC2Provisioner) CreateInstance(opts EC2CreateInstanceOptions) (*cfg.Remote, error) {
	// Set requested region
	p.WithRegion(opts.Region)

	// set highlighter
	var highlight = out.NewColorer(out.CY)

	// Generate authentication
	var keyName = fmt.Sprintf("%s_%s_inertia_key_%d", opts.Name, p.user, time.Now().UnixNano())
	out.Fprintf(p.out, highlight.Sf(":key: Generating key pair '%s'...\n", keyName))
	keyResp, err := p.client.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: aws.String(keyName),
	})
	if err != nil {
		return nil, err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Save key
	var keyPath = filepath.Join(homeDir, ".ssh", *keyResp.KeyName)
	out.Fprintf(p.out, highlight.Sf(":inbox_tray: Saving key to '%s'...\n", keyPath))
	if err = local.SaveKey(*keyResp.KeyMaterial, keyPath); err != nil {
		return nil, err
	}

	// Create security group for network configuration
	var secGroup = fmt.Sprintf("%s-%d", opts.Name, time.Now().UnixNano())
	out.Fprintf(p.out, highlight.Sf(":circus_tent: Creating security group '%s'...\n", secGroup))
	group, err := p.client.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName: aws.String(secGroup),
		Description: aws.String(
			fmt.Sprintf("Rules for project %s on %s", opts.ProjectName, opts.Name),
		),
	})
	if err != nil {
		return nil, err
	}

	// Set rules for ports
	out.Fprintf(p.out, highlight.Sf(":electric_plug: Exposing ports '%s'...\n", secGroup))
	if err = p.exposePorts(*group.GroupId, opts.DaemonPort, opts.Ports); err != nil {
		return nil, err
	}

	// Start up instance
	out.Fprintf(p.out, highlight.Sf(":boat: Requesting instance '%s'...\n", secGroup))
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

	// Check response validity
	if runResp.Instances == nil || len(runResp.Instances) == 0 {
		return nil, errors.New("Unable to start instances: " + runResp.String())
	}
	out.Fprintf(p.out, highlight.Sf("A %s instance has been provisioned", opts.InstanceType))

	// Loop until intance is running
	var instance ec2.Instance
	for {
		// Wait briefly between checks
		time.Sleep(3 * time.Second)

		// Request instance status
		out.Fprintf(p.out, "Checking status of the requested instance...\n")
		result, err := p.client.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{runResp.Instances[0].InstanceId},
		})
		if err != nil {
			return nil, err
		}

		// Check if reservations are present
		if result.Reservations == nil || len(result.Reservations) == 0 ||
			len(result.Reservations[0].Instances) == 0 {
			// A reservation corresponds to a command to start instances
			// If nothing is here... we gotta keep waiting
			fmt.Fprintln(p.out, "No reservations found yet.")
			continue
		}

		// Get status
		s := result.Reservations[0].Instances[0].State
		if s == nil {
			fmt.Println(p.out, "Status unknown.")
			continue
		}

		// Code 16 means instance has started, and we can continue!
		if s.Code != nil && *s.Code == codeEC2InstanceStarted {
			fmt.Fprintln(p.out, "Instance is running!")
			instance = *result.Reservations[0].Instances[0]
			break
		}

		// Otherwise, keep polling
		if s.Name != nil {
			fmt.Fprintln(p.out, "Instance status: "+*s.Name)
		} else {
			fmt.Fprintln(p.out, "Instance status: "+s.String())
		}
		continue
	}

	// Check instance validity
	if instance.PublicDnsName == nil {
		return nil, errors.New("Unable to find public IP address for instance: " + instance.String())
	}

	// Set tags
	out.Fprintf(p.out, "Setting tags on instance...\n")
	if _, err = p.client.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{instance.InstanceId},
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
	}); err != nil {
		fmt.Fprintln(p.out, "Failed to set tags: "+err.Error())
	}

	// Poll for SSH port to open
	fmt.Fprintln(p.out, "Waiting for ports to open...")
	for {
		time.Sleep(3 * time.Second)
		fmt.Fprintln(p.out, "Checking ports...")
		if conn, err := net.Dial("tcp", *instance.PublicDnsName+":22"); err == nil {
			fmt.Fprintln(p.out, "Connection established!")
			conn.Close()
			break
		}
	}

	// Generate webhook secret
	out.Fprintf(p.out, "Generating a webhook secret...\n")
	webhookSecret, err := common.GenerateRandomString()
	if err != nil {
		fmt.Fprintln(p.out, err.Error())
		fmt.Fprintln(p.out, "Using default secret 'inertia'")
		webhookSecret = "interia"
	} else {
		fmt.Fprintf(p.out, "Generated webhook secret: '%s'\n", webhookSecret)
	}

	// Return remote configuration
	return &cfg.Remote{
		Name: opts.Name,
		IP:   *instance.PublicDnsName,
		SSH: &cfg.SSH{
			User:         p.user,
			IdentityFile: keyPath,
			SSHPort:      "22",
		},
		Daemon: &cfg.Daemon{
			Port:          strconv.FormatInt(opts.DaemonPort, 10),
			WebHookSecret: webhookSecret,
		},
		Profiles: make(map[string]string),
	}, nil
}

// WithRegion assigns a region to the client
func (p *EC2Provisioner) WithRegion(region string) {
	p.client.Config.WithRegion(region)
	p.client = ec2.New(p.session, &p.client.Config)
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

func (p *EC2Provisioner) init(user string, creds *credentials.Credentials, out []io.Writer) error {
	if len(out) > 0 {
		p.out = out[0]
	} else {
		p.out = common.DevNull{}
	}
	p.user = user

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
