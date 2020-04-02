/*
Copyright 2019 The Knative Authors

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

package slack

import (
	"context"
	"io/ioutil"
	"path/filepath"

	"github.com/nlopes/slack"

	"github.com/mattmoor/bindings/pkg/bindings"
)

const (
	VolumeName = "slack-binding"
	MountPath  = bindings.MountPath + "/slack" // filepath.Join isn't const.
)

// ReadKey may be used to read keys from the secret bound by the SlackBinding.
func ReadKey(key string) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(MountPath, key))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// AccessToken reads the file named accessToken that is mounted by the SlackBinding.
func AccessToken() (string, error) {
	return ReadKey("token")
}

// New instantiates a new github client from the access token from the SlackBinding
func New(ctx context.Context) (*slack.Client, error) {
	at, err := AccessToken()
	if err != nil {
		return nil, err
	}
	return slack.New(at), nil
}
