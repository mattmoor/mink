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

package twitter

import (
	"context"
	"io/ioutil"
	"path/filepath"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/mattmoor/bindings/pkg/bindings"
)

const (
	VolumeName = "twitter-binding"
	MountPath  = bindings.MountPath + "/twitter" // filepath.Join isn't const.
)

// ReadKey may be used to read keys from the secret bound by the TwitterBinding.
func ReadKey(key string) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(MountPath, key))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// NewAppClient instantiates a new twitter client that authenticates with the application credentials.
func NewAppClient(ctx context.Context) (*twitter.Client, error) {
	consumerKey, err := ReadKey("consumerKey")
	if err != nil {
		return nil, err
	}
	consumerSecretKey, err := ReadKey("consumerSecretKey")
	if err != nil {
		return nil, err
	}
	// oauth2 configures a client that uses app credentials to keep a fresh token
	config := &clientcredentials.Config{
		ClientID:     consumerKey,
		ClientSecret: consumerSecretKey,
		TokenURL:     "https://api.twitter.com/oauth2/token",
	}
	// http.Client will automatically authorize Requests
	httpClient := config.Client(oauth2.NoContext)

	// Twitter client
	return twitter.NewClient(httpClient), nil
}

// NewUserClient instantiates a new twitter client that authenticates with a particular
// user access token and secret in addition to the application credentials.
func NewUserClient(ctx context.Context) (*twitter.Client, error) {
	consumerKey, err := ReadKey("consumerKey")
	if err != nil {
		return nil, err
	}
	consumerSecretKey, err := ReadKey("consumerSecretKey")
	if err != nil {
		return nil, err
	}
	accessToken, err := ReadKey("accessToken")
	if err != nil {
		return nil, err
	}
	accessSecret, err := ReadKey("accessSecret")
	if err != nil {
		return nil, err
	}
	config := oauth1.NewConfig(consumerKey, consumerSecretKey)
	token := oauth1.NewToken(accessToken, accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	// Twitter client
	return twitter.NewClient(httpClient), nil
}
