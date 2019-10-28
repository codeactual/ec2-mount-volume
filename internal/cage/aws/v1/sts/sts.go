// Copyright (C) 2019 The CodeActual Go Environment Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sts

import (
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/pkg/errors"
)

// NewBasicEC2RoleProvider returns an EC2RoleProvider for a given role.
func NewBasicEC2RoleProvider() (*ec2rolecreds.EC2RoleProvider, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(sess)}, nil
}
