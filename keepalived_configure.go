package main

import (
	"./gen"
	"./vip/detali"
	"flag"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"os"
)

func init() {
	flag.Parse()
}

func main() {
	gen.CheckArg()
	cfg := aws.Config{
		Region: aws.String(*gen.Region),
	}
	svc := ec2.New(session.New(&cfg))

	if vipvalue, state, red, err := gen.GetAlived(svc); !detail.OutError(err) {
		os.Exit(1)
	} else {
		gen.GenerateConfig(&vipvalue, &state, &red)
	}
}
