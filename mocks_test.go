package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type mock struct {
	MockNewResource func(*pulumi.MockResourceArgs)
	MockCall        func(pulumi.MockCallArgs) (*resource.PropertyMap, error)
}

func (m mock) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	if m.MockNewResource != nil {
		// Even if this function returns an error, it doesn't resuult in the original
		// function call returning an error.  There is no way to introduce a failure
		// because this is being run as a goroutine in the pulumi.Ctx.
		m.MockNewResource(&args)
	}

	return args.Name + "_id", args.Inputs, nil
}

func (m mock) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	if m.MockCall != nil {
		outputs, err := m.MockCall(args)
		if err != nil {
			return resource.PropertyMap{}, err
		}

		if outputs != nil {
			return *outputs, nil
		}
	}

	return args.Args, nil
}

// Solution from: https://github.com/pulumi/pulumi/issues/4472#issuecomment-731012097

// WithMocksAndConfig is a duplicate of the pulumi.WithMocks() function with the
// additional configuration value of 'config'.
func WithMocksAndConfig(project, stack string, config map[string]string, mocks pulumi.MockResourceMonitor) pulumi.RunOption {
	return func(info *pulumi.RunInfo) {
		info.Project, info.Stack, info.Mocks, info.Config = project, stack, mocks, config
	}
}
