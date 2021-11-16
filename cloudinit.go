package main

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// cloud init file pieces, with placeholders for things that change.
const (
	writeDirective = `
write_files:
    string_1: %s
    string_2: %s
`

	runDirective = `
runcmd:
  - hostnamectl --no-ask-password set-hostname %s
  - thing_1='%s' thing_2='%s' /bin/program
  - systemctl reboot
`
)

type Supplier struct {
	token pulumi.StringOutput
	value string
}

// NewSupplier makes a new Supplier that will create the cloud-init payload
// we'll send to the VM later.
func NewSupplier(ctx *pulumi.Context, service string) Supplier {
	c := config.New(ctx, service)
	s := Supplier{
		token: c.RequireSecret("thing.token"),
		value: c.Require("thing.address"),
	}

	return s
}

// Compose takes in the hostname and returns the cloud-init directives to run on
// instance startup.
func (s Supplier) Compose(region, hostname string) pulumi.StringOutput {
	payload := pulumi.Unsecret(s.token).ApplyT(func(token string) string {
		var output bytes.Buffer

		output.Write([]byte("#cloud-config\n"))
		fmt.Fprintf(&output, writeDirective, hostname, region)
		fmt.Fprintf(&output, runDirective, hostname, token, s.value)

		// Base64 encode the string to ensure it does not get changed by accident.
		return base64.StdEncoding.EncodeToString(output.Bytes())
	}).(pulumi.StringOutput)

	// Put file into pulumi secret
	secret := pulumi.ToSecret(payload)

	return secret.(pulumi.StringOutput)
}
