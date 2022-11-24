package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/transport/http"
	"github.com/jessevdk/go-flags"
	"os"
)

func main() {

	var opts struct {
		Type        string `short:"t" long:"type" description:"Cloud type" choice:"aws" choice:"azure" choice:"gcp" choice:"ali" default:"aws" required:"true"`
		Mode        string `short:"m" long:"mode" description:"Mode of operation" choice:"register" choice:"unregister" default:"register" required:"true"`
		TargetGroup string `short:"g" long:"group" description:"[AWS] Target group ARN to register at / unregister from"`
		Instance    string `short:"i" long:"instance" description:"Instance id to register / unregister"`
	}

	_, err := flags.ParseArgs(&opts, os.Args)

	if err != nil {
		os.Exit(1)
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1") /*config.WithClientLogMode(aws.LogRequestWithBody)*/)
	if err != nil {
		panic(err)
	}

	tgtGroup := "arn:aws:elasticloadbalancing:us-east-1:205379741905:targetgroup/cf-proxy-aws-cfn04-tg-https/5629e6c89ae42766"
	tgtId := "i-0e77c77f4e17f7b00"
	tgtDesc := []types.TargetDescription{{Id: &tgtId}}

	lbClient := elasticloadbalancingv2.NewFromConfig(cfg)

	switch opts.Mode {
	case "register":
		fmt.Println("Registering", tgtId, "at target group", tgtGroup)

		rti := &elasticloadbalancingv2.RegisterTargetsInput{
			TargetGroupArn: &tgtGroup,
			Targets:        tgtDesc,
		}

		rto, err := lbClient.RegisterTargets(ctx, rti)
		if err != nil {
			fmt.Println("Could not register", tgtId, "at target group", tgtGroup)
			fmt.Println(err)
			os.Exit(1)
		}
		response := middleware.GetRawResponse(rto.ResultMetadata).(*http.Response)
		if response != nil {
			fmt.Println(response.Status)
		}

	case "unregister":
		fmt.Println("Unregistering", tgtId, "from target group", tgtGroup)

		dti := &elasticloadbalancingv2.DeregisterTargetsInput{
			TargetGroupArn: &tgtGroup,
			Targets:        tgtDesc,
		}

		dto, err := lbClient.DeregisterTargets(ctx, dti)
		if err != nil {
			fmt.Println("Could not unregister", tgtId, "from target group", tgtGroup)
			fmt.Println(err)
			os.Exit(1)
		}
		response := middleware.GetRawResponse(dto.ResultMetadata).(*http.Response)
		if response != nil {
			fmt.Println(response.Status)
		}
	}
}

type ErrCoder interface {
	ErrorCode() string
}

// TODO: Helper needed to unwrap S3 errors due to https://github.com/aws/aws-sdk-go-v2/issues/1386
func IsAwsError(err error, tgt ErrCoder) bool {
	if err == nil {
		return false
	}
	var apiErr smithy.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.ErrorCode() == tgt.ErrorCode()
}
