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

package resources

import (
	"fmt"
	"strings"

	"github.com/vaikas/postgressource/pkg/apis/sources/v1alpha1"
	"knative.dev/pkg/kmeta"
)

const (
	createFunction = `
CREATE OR REPLACE FUNCTION POSTGRES_SOURCE_NAME() RETURNS TRIGGER AS $$

    DECLARE 
        data json;
        notification json;
    
    BEGIN
    
        -- Convert the old or new row to JSON, based on the kind of action.
        -- Action = DELETE?             -> OLD row
        -- Action = INSERT or UPDATE?   -> NEW row
        IF (TG_OP = 'DELETE') THEN
            data = row_to_json(OLD);
        ELSE
            data = row_to_json(NEW);
        END IF;
        
        -- Contruct the notification as a JSON string.
        notification = json_build_object(
                          'table',TG_TABLE_NAME,
                          'action', TG_OP,
                          'data', data);
        
                        
        -- Execute pg_notify(channel, notification)
        PERFORM pg_notify('POSTGRES_SOURCE_NAME',notification::text);
        
        -- Result is ignored since this is an AFTER trigger
        RETURN NULL; 
    END;
    
$$ LANGUAGE plpgsql;
`

	createTrigger = `
CREATE TRIGGER POSTGRES_SOURCE_NAME
AFTER INSERT OR UPDATE OR DELETE ON POSTGRES_SOURCE_TABLE
    FOR EACH ROW EXECUTE PROCEDURE POSTGRES_SOURCE_NAME();
`

	dropTrigger  = `DROP TRIGGER IF EXISTS POSTGRES_SOURCE_NAME ON POSTGRES_SOURCE_TABLE;`
	dropFunction = `DROP FUNCTION IF EXISTS POSTGRES_SOURCE_NAME;`

	GetTriggersQuery = `select trigger_name, event_manipulation, event_object_table from information_schema.triggers where event_object_table =$1 and trigger_name =$2`

	GetFunctionQuery = `select proname as function_name from pg_proc where proname = $1`

	GetTableQuery = `select tablename from pg_catalog.pg_tables where tablename = $1`
)

// Make postgres compatible name just like we do for k8s (<=63 chars)
// and convert all the - into underscores.
func MakePostgresName(source *v1alpha1.PostgresSource) string {
	return strings.ReplaceAll(kmeta.ChildName(fmt.Sprintf("postgressource-%s-", source.Name), string(source.GetUID())), "-", "_")

}

func MakeFunction(source *v1alpha1.PostgresSource) string {
	return strings.ReplaceAll(createFunction, "POSTGRES_SOURCE_NAME", MakePostgresName(source))
}

func MakeDropFunction(source *v1alpha1.PostgresSource) string {
	return strings.ReplaceAll(dropFunction, "POSTGRES_SOURCE_NAME", MakePostgresName(source))
}

func MakeTrigger(source *v1alpha1.PostgresSource, table string) string {
	return strings.ReplaceAll(strings.ReplaceAll(createTrigger, "POSTGRES_SOURCE_NAME", MakePostgresName(source)), "POSTGRES_SOURCE_TABLE", table)
}

func MakeDropTrigger(source *v1alpha1.PostgresSource, table string) string {
	return strings.ReplaceAll(strings.ReplaceAll(dropTrigger, "POSTGRES_SOURCE_NAME", MakePostgresName(source)), "POSTGRES_SOURCE_TABLE", table)
}
