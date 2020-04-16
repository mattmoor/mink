# Knative Eventing Source for Postgres

Knative Eventing `postgressource` defines a simple Postgres source. This lets
you to create a PostgresSource and define the tables that you would like to be
notified of changes to (insert/update/delete) and it will send `Cloud Event` for
each modification to those tables.

## Installation

To install Postgres Source, decide which namespace you want to run the
controllers in and modify the config/* files appropriately. By default it
installs into `knative-sources` namespace.

```shell
ko apply -f ./config/
```

### SHOW ME

Ok, so once you get everything up and running, then you can create a Postgres
Source binding. Before you can do that, you need to create a `Secret` that has
the credentials for accessing your Postgres database. The secret must have the
connection string used to connect to the Postgres database and the field must be
called `connectionstr`.


Say my db is at `127.0.0.1` and my username is `foobar` and password is `really`
and the database I want to use is `users`, I could create that secret like so:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sql-secret
  namespace: default
stringData:
  connectionstr: postgres://foobar:really@127.0.0.1:5432/users
```

Other ways to specify the connectionstr that postgres expects for its open
connection should be supported, but has not been tested...

Furthermore, say the table I want to get notified on when things change is
called `orders` and I want to send those events to my default `Broker` in the
default namespace, I could create a PostgresSource like so:

```yaml
apiVersion: sources.vaikas.dev/v1alpha1
kind: PostgresSource
metadata:
  name: vaikaspostgres
  namespace: default
spec:
  tables:
  - name: orders
  secret:
    name: sql-secret
  sink:
    ref:
      apiVersion: eventing.knative.dev/v1beta1
      kind: Broker
      name: default
```

And that's it. Once you create that, any modifications to orders table are now
fed as Cloud Events into the default Broker and then you can use Knative
Eventing `Trigger`s to process those events as you see fit.

## Inner workings

### Function

The Source creates a Postgres Function that looks like so. We create one
function per source, and name it uniquely and then reuse it in all the Triggers
we create for the tables that we create notifications on. So, the
POSTGRES_SOURCE_NAME will be changed to something like:
postgressource_<your-postgres-sourcename>_<uuid of the source> that then gets
truncated to 63 characters which is the maximum length postgres by default
handles. This is the same naming convention we use for the k8s resources we
create, but instead of `-` we use `_` because it's not allowed in the postgres
names. Also note that we create a pg_notify channel that's named the same so
that the `Receive Adapter` can watch for for notifications for and turn them
into Cloud Events.

```
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

```

### Triggers

Then for each of the Tables that you specify in `spec.tables`, we create a
trigger that will call that function. That Trigger looks like so, but again,
replace the `POSTGRES_SOURCE_NAME` with the name of the postgres resource we
construct as per above, and the `POSTGRES_SOURCE_TABLE` with the table in
`spec.tables`.

```
CREATE TRIGGER POSTGRES_SOURCE_NAME
AFTER INSERT OR UPDATE OR DELETE ON POSTGRES_SOURCE_TABLE
    FOR EACH ROW EXECUTE PROCEDURE POSTGRES_SOURCE_NAME();

```

### Receive Adapter / SQLBinding

For the dataplane, we have to watch for those notifications, so we create a
Deployment that creates a `Listen` against the notification channel for our
events and sends them to the specified `Sink`. By default, we also create a
dedicated `Service Account` in the namespace the Postgres Source is created and
create a Role Binding for being able to use `ConfigMap`s for getting logging
config information as well as to (future work) be able to (possibly) checkpoint
our work and hence not miss any events in case of failures.

To learn more about Knative, please visit our
[Knative docs](https://github.com/knative/docs) repository.

If you are interested in contributing, see [CONTRIBUTING.md](./CONTRIBUTING.md)
and [DEVELOPMENT.md](./DEVELOPMENT.md).
