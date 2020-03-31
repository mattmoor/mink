/*
Copyright 2020 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vsphere

import (
	"context"
	"io/ioutil"
	"net/url"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25/soap"
	corev1 "k8s.io/api/core/v1"
)

const (
	VolumeName = "vsphere-binding"
	MountPath  = "/var/bindings/vsphere" // filepath.Join isn't const.
)

type EnvConfig struct {
	Insecure bool   `envconfig:"GOVC_INSECURE" default:"false"`
	Address  string `envconfig:"GOVC_URL" required:"true"`
}

// ReadKey may be used to read keys from the secret.
func ReadKey(key string) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(MountPath, key))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func New(ctx context.Context) (*govmomi.Client, error) {
	var env EnvConfig
	if err := envconfig.Process("", &env); err != nil {
		return nil, err
	}

	parsedURL, err := soap.ParseURL(env.Address)
	if err != nil {
		return nil, err
	}

	// Read the username and password from the filesystem.
	username, err := ReadKey(corev1.BasicAuthUsernameKey)
	if err != nil {
		return nil, err
	}
	password, err := ReadKey(corev1.BasicAuthPasswordKey)
	if err != nil {
		return nil, err
	}
	parsedURL.User = url.UserPassword(username, password)

	return govmomi.NewClient(ctx, parsedURL, env.Insecure)
}

func NewREST(ctx context.Context) (*rest.Client, error) {
	var env EnvConfig
	if err := envconfig.Process("", &env); err != nil {
		return nil, err
	}

	parsedURL, err := soap.ParseURL(env.Address)
	if err != nil {
		return nil, err
	}

	// Read the username and password from the filesystem.
	username, err := ReadKey(corev1.BasicAuthUsernameKey)
	if err != nil {
		return nil, err
	}
	password, err := ReadKey(corev1.BasicAuthPasswordKey)
	if err != nil {
		return nil, err
	}
	parsedURL.User = url.UserPassword(username, password)

	soapclient, err := govmomi.NewClient(ctx, parsedURL, env.Insecure)
	if err != nil {
		return nil, err
	}

	// For whatever reason the rest client doesn't inherit the SOAP client's auth.
	restclient := rest.NewClient(soapclient.Client)
	if err := restclient.Login(ctx, parsedURL.User); err != nil {
		return nil, err
	}
	return restclient, nil
}

func Address(ctx context.Context) (string, error) {
	var env EnvConfig
	if err := envconfig.Process("", &env); err != nil {
		return "", err
	}

	return env.Address, nil
}
