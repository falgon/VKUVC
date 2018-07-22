package detail

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"strings"
)

type ec2Filter []*ec2.Filter

func (fl *ec2Filter) Push(key, value string) {
	for _, f := range *fl {
		if f.Name != nil && *f.Name == key {
			f.Values = append(f.Values, &value)
			return
		}
	}
	*fl = append(*fl, &ec2.Filter{Name: &key, Values: []*string{&value}})
}

func NewFilter() ec2Filter {
	return make(ec2Filter, 0)
}

func GetThisInstanceId() (r string, err error) {
	svc := ec2metadata.New(session.Must(session.NewSession()))
	var doc ec2metadata.EC2InstanceIdentityDocument
	if doc, err = svc.GetInstanceIdentityDocument(); err == nil {
		r = doc.InstanceID
	}
	return
}

func OutError(err error) bool {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return false
	}
	return true
}

func ParseKeyValue(s *string) (key, value string, err error) {
	b := true
	for _, t := range strings.Split(*s, ",") {
		section := strings.Split(t, "=")
		switch section[0] {
		case "Key":
			key = section[1]
		case "Value":
			value = section[1]
		default:
			b = false
		}
	}
	if !b {
		err = errors.New("Invalid statements.")
		return
	}
	return
}

func GetRedudancies(svc *ec2.EC2, red *string) (*ec2.DescribeInstancesOutput, error) {
	if name, value, err := ParseKeyValue(red); err != nil {
		return nil, err
	} else {
		input := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("tag:" + name),
					Values: []*string{
						aws.String(value),
					},
				},
			},
		}
		return svc.DescribeInstances(input)
	}
}
