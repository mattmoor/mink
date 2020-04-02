# Knative-style Bindings

This repository contains a collection of Knative-style Bindings for accessing
various services.  Each of the bindings contained in this repository generally
has two key parts:

1. A Binding CRD that will augment the runtime contract of the Binding subject
   in some way.

2. A small client library for consuming the runtime conract alteration to
   bootstrap and API client for the service being bound.


## `GithubBinding`

The `GithubBinding` is intended to facilitate the consumption of the GitHub API.
It has the following form:

```yaml
apiVersion: bindings.mattmoor.dev/v1alpha1
kind: GithubBinding
metadata:
  name: foo-binding
spec:
  subject:
    apiVersion: apps/v1
    kind: Deployment
    # Either name or selector may be specified.
    selector:
      matchLabels:
        foo: bar

  secret:
    name: github-secret
```

The referenced secret should have a key named `accessToken` with the Github
access token to be used with the Github API.  It and any other keys are made
available under `/var/bindings/github/` (this is the runtime contract of the
`GithubBinding`).

There is a helper library available to aid in the consumption of this runtime
contract, which returns a `github.com/google/go-github/github.Client`:

```go

import "github.com/mattmoor/bindings/pkg/github"


// Instantiate a Client from the access token made available by
// the GithubBinding.
client, err := github.New(ctx)
...

```


## `SlackBinding`

The `SlackBinding` is intended to facilitate the consumption of the Slack API.
It has the following form:

```yaml
apiVersion: bindings.mattmoor.dev/v1alpha1
kind: SlackBinding
metadata:
  name: foo-binding
spec:
  subject:
    apiVersion: apps/v1
    kind: Deployment
    # Either name or selector may be specified.
    selector:
      matchLabels:
        foo: bar

  secret:
    name: slack-secret
```

The referenced secret should have a key named `token` with the Slack
token to be used with the Slack API.  It and any other keys are made
available under `/var/bindings/slack/` (this is the runtime contract of the
`SlackBinding`).

There is a helper library available to aid in the consumption of this runtime
contract, which returns a `github.com/nlopes/slack.Client`:

```go

import "github.com/mattmoor/bindings/pkg/slack"


// Instantiate a Client from the token made available by
// the SlackBinding.
client, err := slack.New(ctx)
...

```


## `TwitterBinding`

The `TwitterBinding` is intended to facilitate the consumption of the Twitter API.
It has the following form:

```yaml
apiVersion: bindings.mattmoor.dev/v1alpha1
kind: TwitterBinding
metadata:
  name: foo-binding
spec:
  subject:
    apiVersion: apps/v1
    kind: Deployment
    # Either name or selector may be specified.
    selector:
      matchLabels:
        foo: bar

  secret:
    name: twitter-secret
```

The referenced secret must have the keys: `consumerKey`, `consumerSecretKey`,
which are the Twitter "Application" credentials.  It may also optionally
have the keys: `accessToken`, `accessSecret` in order to access the Twitter
API using "User" credentials.  These (and other) keys are made available
under `/var/bindings/twitter/` (this is the runtime contract of the
`TwitterBinding`).

Depending on whether you want "Application" or "User" functionality (the latter
requires additional secret keys), we provide a helper for each to instantiate
a client compatible with `github.com/dghubble/go-twitter`:

```go

import "github.com/mattmoor/bindings/pkg/twitter"


// Instantiate a Client that authenticates as an Application from
// the TwitterBinding.
client, err := twitter.NewAppClient(ctx)
...

// Instantiate a Client that authenticates as a User from
// the TwitterBinding.
client, err := twitter.NewUserClient(ctx)
...

```

## `GoogleCloudSQLBinding`

The `GoogleCloudSQLBinding` is intended to facilitate the consumption of Google
Cloud SQL instances via
[the proxy](https://github.com/GoogleCloudPlatform/cloudsql-proxy) without
manually configuring it.  It has the following form:

```yaml
apiVersion: bindings.mattmoor.dev/v1alpha1
kind: GoogleCloudSQLBinding
metadata:
  name: foo-binding
spec:
  subject:
    apiVersion: apps/v1
    kind: Deployment
    # Either name or selector may be specified.
    selector:
      matchLabels:
        foo: bar

  secret:
    name: cloudsql-secret

  instance: "project:region:name"
```

The referenced secret should have three parts:
1. `credentials.json`: the JSON Key with the Cloud SQL "Client" IAM role.
2. `username`: the database username to use when connecting.
3. `password`: the database password to use when connecting.

These keys are made available under `/var/bindings/cloudsql/secrets/`, and a unix
socket to the instance if made available under `/var/bindings/cloudsql/sockets/`.

There is a helper library available to facilitate consumption a `database/sql.DB`:

```go

import "github.com/mattmoor/bindings/pkg/cloudsql"

// Open a connection to the named database.
db, err := cloudsql.Open(ctx, "DATABASE NAME")
...

```

## `SQLBinding`

The `SQLBinding` is intended to facilitate the consumption of SQL instances.

```yaml
apiVersion: bindings.mattmoor.dev/v1alpha1
kind: SQLBinding
metadata:
  name: foo-binding
spec:
  subject:
    apiVersion: apps/v1
    kind: Deployment
    # Either name or selector may be specified.
    selector:
      matchLabels:
        foo: bar

  secret:
    name: sql-secret
```

The referenced secret should have the connection string in it:
1. `connectionstr`: in the format expected by the golang SQL package. Omit the database.

For example, to connect to postgres you would specify
postgres://username:password@ip:port/

postgres://myuser:mysupersecretpassword@127.0.0.1:5432/

This key is made available under `/var/bindings/sql/secrets/`

There is a helper library available to facilitate consumption a `database/sql.DB`:

```go

import (
	"github.com/mattmoor/bindings/pkg/sql"

	// Also import your database driver, for example for postgres:
	_ "github.com/lib/pq"
)

// Open a connection to the named database.
db, err := cloudsql.Open(ctx, "postgres", "DATABASENAME")
...

```
