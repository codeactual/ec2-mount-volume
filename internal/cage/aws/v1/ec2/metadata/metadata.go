// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package metadata

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/pkg/errors"

	cage_sts "github.com/codeactual/ec2-mount-volume/internal/cage/aws/v1/sts"
)

func Service() (*ec2metadata.EC2Metadata, error) {
	s, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	return ec2metadata.New(s), nil
}

func Session() (*session.Session, error) {
	metaSvc, err := Service()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	idDoc, err := metaSvc.GetInstanceIdentityDocument()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	p, err := cage_sts.NewBasicEC2RoleProvider()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	creds := credentials.NewCredentials(p)
	config := &aws.Config{Credentials: creds, Region: aws.String(idDoc.Region)}
	return session.NewSession(config)
}

func Volumes() (volumes []*ec2.Volume, err error) {
	sess, err := Session()
	if err != nil {
		return []*ec2.Volume{}, errors.WithStack(err)
	}

	ec2svc := ec2.New(sess)

	metaSvc, err := Service()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	idDoc, err := metaSvc.GetInstanceIdentityDocument()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	input := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.instance-id"),
				Values: []*string{aws.String(idDoc.InstanceID)},
			},
		},
	}

	for {
		output, err := ec2svc.DescribeVolumes(input)
		if err != nil {
			return []*ec2.Volume{}, errors.WithStack(err)
		}

		volumes = append(volumes, output.Volumes...)

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return volumes, nil
}
