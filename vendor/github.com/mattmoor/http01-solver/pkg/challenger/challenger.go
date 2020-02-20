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

package challenger

import (
	context "context"
	"net/http"
	"sync"
)

// Interface defines the interface for handling register, unregistering,
// and serving challenge responses.
type Interface interface {
	http.Handler

	RegisterChallenge(path, response string)
	UnregisterChallenge(path string)
}

func New(ctx context.Context) (Interface, error) {
	return &challenger{}, nil
}

type challenger struct {
	sync.RWMutex

	paths map[string]string
}

var _ Interface = (*challenger)(nil)

func (c *challenger) RegisterChallenge(path, response string) {
	c.Lock()
	defer c.Unlock()

	if c.paths == nil {
		c.paths = make(map[string]string, 1)
	}
	c.paths[path] = response
}

func (c *challenger) UnregisterChallenge(path string) {
	c.Lock()
	defer c.Unlock()

	if c.paths != nil {
		delete(c.paths, path)
	}
}

func (c *challenger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.RLock()
	defer c.RUnlock()

	if c.paths == nil {
		http.Error(w, "Unknown path", http.StatusNotFound)
		return
	}
	resp, ok := c.paths[r.URL.Path]
	if !ok {
		http.Error(w, "Unknown path", http.StatusNotFound)
		return
	}
	w.Write([]byte(resp))
}
