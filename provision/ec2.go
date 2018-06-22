package provision

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// EC2Provisioner creates Amazon EC2 instances
type EC2Provisioner struct {
	client *ec2.EC2
}

func NewEC2Provisioner() *EC2Provisioner {
	sess := session.Must(session.NewSession())
	return &EC2Provisioner{
		client: ec2.New(sess, &aws.Config{}),
	}
}

func (p *EC2Provisioner) CreateInstance(instanceType, region string) (string, error) {
	quantity := int64(1)
	out, err := p.client.AllocateHosts(&ec2.AllocateHostsInput{
		AvailabilityZone: &region,
		InstanceType:     &instanceType,
		Quantity:         &quantity,
	})
	if err != nil {
		return "", err
	}
	if len(out.HostIds) == 0 {
		return "", errors.New("no host created: " + out.String())
	}
	return *out.HostIds[0], nil
}
