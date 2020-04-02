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

package sql

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
	VolumeName = "sql-binding"
	MountPath  = bindings.MountPath + "/sql" // filepath.Join isn't const.
)

// ReadKey may be used to read keys from the secret bound by the SQLBinding.
func ReadKey(key string) (string, error) {
	data, err := ioutil.ReadFile(filepath.Join(MountPath, key))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Open returns a *sql.DB that has been authenticated with the named database.
func Open(ctx context.Context, dbType, database string) (*sql.DB, error) {
	// TODO: Add more free form strings here... Like, allow for Username, Password
	// etc. etc.
	/*
		username, err := ReadKey("username")
		if err != nil {
			return nil, err
		}
		password, err := ReadKey("password")
		if err != nil {
			return nil, err
		}
	*/
	connStr, err := ReadKey("connectionstr")
	if err != nil {
		return nil, err
	}
	return sql.Open(dbType, fmt.Sprintf("%s/%s", connStr, database))
}
