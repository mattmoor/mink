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

package cloudsql

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"

	"github.com/mattmoor/bindings/pkg/bindings"
)

const (
	ContainerName    = "cloudsql-proxy"
	SecretVolumeName = "cloudsql-credentials-binding"
	SocketVolumeName = "cloudsql-socket-binding"
	MountPath        = bindings.MountPath + "/cloudsql" // filepath.Join isn't const.
	SecretMountPath  = MountPath + "/secrets"           // filepath.Join isn't const.
	SocketMountPath  = MountPath + "/socket"            // filepath.Join isn't const.
)

// ReadKey may be used to read keys from the secret bound by the GoogleCloudSQLBinding.
func ReadKey(key string) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(SecretMountPath, key))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Open returns a *sql.DB that has been authenticated with the named database.
func Open(ctx context.Context, database string) (*sql.DB, error) {
	username, err := ReadKey("username")
	if err != nil {
		return nil, err
	}
	password, err := ReadKey("password")
	if err != nil {
		return nil, err
	}
	return sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", username, password, database))
}
