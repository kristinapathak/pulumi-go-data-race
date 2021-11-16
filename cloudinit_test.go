package main

import (
	"encoding/base64"
	"sync"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	expectedTestOne = `#cloud-config

write_files:
    string_1: greatest-host-ever
    string_2: region-1

runcmd:
  - hostnamectl --no-ask-password set-hostname greatest-host-ever
  - thing_1='secret_secret' thing_2='http://example.com' /bin/program
  - systemctl reboot
`
	expectedTestTwo = `#cloud-config

write_files:
    string_1: slowestHost
    string_2: SecondRegion

runcmd:
  - hostnamectl --no-ask-password set-hostname slowestHost
  - thing_1='bad secret' thing_2='1060 West Addison Street' /bin/program
  - systemctl reboot
`
)

func TestNewSupplier(t *testing.T) {
	config := map[string]string{
		"project:thing.address": "http://example.com",
		"project:thing.token":   "secret_secret",
	}

	tests := []struct {
		description string
		service     string
		config      map[string]string
	}{
		{
			description: "Success",
			service:     "project",
			config:      config,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			pulumi.RunErr(func(ctx *pulumi.Context) error {
				ci := NewSupplier(ctx, tc.service)

				assert.NotNil(ci)
				return nil
			}, WithMocksAndConfig("project", "stack", tc.config,
				mock{
					MockNewResource: func(args *pulumi.MockResourceArgs) {
						pp.Print(args)
					},
					MockCall: func(args pulumi.MockCallArgs) (*resource.PropertyMap, error) {
						pp.Print(args)
						return nil, nil
					},
				},
			))
		})
	}
}

func TestCompose(t *testing.T) {
	tests := []struct {
		description string
		service     string
		config      map[string]string
		region      string
		hostname    string
		expect      string
	}{
		{
			description: "Success with one set of strings",
			service:     "project",
			region:      "region-1",
			hostname:    "greatest-host-ever",
			expect:      expectedTestOne,
			config: map[string]string{
				"project:thing.address": "http://example.com",
				"project:thing.token":   "secret_secret",
			},
		},
		{
			description: "Success with a different set of strings",
			service:     "sample",
			region:      "SecondRegion",
			hostname:    "slowestHost",
			expect:      expectedTestTwo,
			config: map[string]string{
				"sample:thing.address": "1060 West Addison Street",
				"sample:thing.token":   "bad secret",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				ci := NewSupplier(ctx, tc.service)

				require.NotNil(t, ci)

				out := ci.Compose(tc.region, tc.hostname)
				var wg sync.WaitGroup
				wg.Add(1)

				out.ApplyT(func(payload string) string {
					require.True(t, len(payload) > 0)

					decoded, err := base64.URLEncoding.DecodeString(payload)
					assert.NoError(err)
					assert.Equal(tc.expect, string(decoded))

					wg.Done()
					return ""
				})

				wg.Wait()
				return nil
			}, WithMocksAndConfig("project", "stack", tc.config,
				mock{
					MockNewResource: func(args *pulumi.MockResourceArgs) {
						pp.Print(args)
					},
					MockCall: func(args pulumi.MockCallArgs) (*resource.PropertyMap, error) {
						pp.Print(args)
						return nil, nil
					},
				},
			))
			assert.NoError(err)
		})
	}
}
