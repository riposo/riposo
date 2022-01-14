# Riposo Server

Riposo is a JSON document store written in [Go](https://golang.org), based on
the design of [Kinto](https://docs.kinto-storage.org/en/latest/index.html). It
implements [v1.22](https://docs.kinto-storage.org/en/latest/api/index.html) of
the Kinto API and aims to be fully compatible with the client libraries
[kinto-http.js](https://github.com/Kinto/kinto-http.js),
[kinto-http.py](https://github.com/Kinto/kinto-http.py) as well as higher-level,
offline-first abstraction [kinto.js](https://github.com/Kinto/kinto.js).

## Getting Started

The easiest way to deploy Riposo is via Docker.

To launch a test instance (on port 8888):

```shell
docker run --rm -p 8888:8888 riposo/riposo server
```

You can use environment variables to configure all aspects of your instance,
including persistence. For more information, please see the
[Configuration](#configuration) section.

```shell
docker run --rm \
    -e RIPOSO_STORAGE_URL=postgres://postgres:postgres@database/riposo?timezone=UTC \
    -e RIPOSO_PERMISSION_URL=postgres://postgres:postgres@database/riposo?timezone=UTC \
    -e RIPOSO_CACHE_URL=postgres://postgres:postgres@database/riposo?timezone=UTC \
    riposo/riposo:0.1.0 server
```

Plugins can also be loaded dynamically by referencing the plugin ID. In order to
enable a plugin, the server binary must have been built with it. For more
information, please see the [Plugins](#plugins) section.

```shell
# list plugins
docker run --rm riposo/riposo plugins

# run server with plugins
docker run --rm \
    -v $(pwd):/data
    -e RIPOSO_PLUGINS=accounts,default-bucket \
    -e RIPOSO_PERMISSION_DEFAULTS='{"account:create":[system.Everyone], "bucket:create":[system.Authenticated]}' \
    riposo/riposo server
```

## Configuration

The server instance can be configured via environment variables or an optional
[YAML](https://yaml.org/) configuration file which can be passed via `-config`
flag.

| Option                      | Type                   | Description                                                                         | Default                             |
| --------------------------- | ---------------------- | ----------------------------------------------------------------------------------- | ----------------------------------- |
| `project.name`              | `string`               | Project name, as reported by `GET /v1/`                                             | `riposo`                            |
| `project.version`           | `string`               | Project version, as reported by `GET /v1/`                                          | _none_                              |
| `project.docs`              | `string`               | Project documentation URL                                                           | `https://github.com/riposo/riposo/` |
| `id.factory`                | `string`               | ID generator, see [ID Factories](#id-factories)                                     | `nanoid`                            |
| `storage.url`               | `string`               | Storage ba ckend URL, see [Data Backends](#data-backends)                           | `:memory:`                          |
| `permission.url`            | `string`               | Permission backend URL                                                              | `:memory:`                          |
| `permission.defaults`       | `map<string,string[]>` | Default permissions                                                                 | _none_                              |
| `cache.url`                 | `string`               | Cache back end URL                                                                  | `:memory:`                          |
| `batch.max_requests`        | `int`                  | Maximum permitted number of requests per batch                                      | `25`                                |
| `auth.methods`              | `string[]`             | Comma-separated list of auth methods, see [Authentication](#authentication)         | `basic`                             |
| `auth.hash`                 | `string`               | Hash method used for password hashing, `argon2id` or `bcrypt`                       | `argon2id`                          |
| `cors.origins`              | `string[]`             | Permitted CORS origins                                                              | `*`                                 |
| `cors.max_age`              | `duration`             | Indicates how long the results of a preflight CORS request can be cached            | `1h`                                |
| `pagination.token_validity` | `duration`             | Pagination token TTL                                                                | `10m`                               |
| `pagination.max_limit`      | `int`                  | Maximum number of records that can be requested per page                            | `10000`                             |
| `backoff.duration`          | `duration`             | Provide clients with a backoff during which they should avoid unnecessary requests. | _none_                              |
| `backoff.percentage`        | `int`                  | Send backoff header to a fraction of clients.                                       | _none_                              |
| `retry_after`               | `duration`             | Duration after which the client should issue requests after failures.               | `30s`                               |
| `server.address`            | `string`               | Server address to listen for requests                                               | `:8888`                             |
| `server.read_timeout`       | `duration`             | Read timeout for server requests                                                    | `60s`                               |
| `server.write_timeout`      | `duration`             | Write timeout for server responses                                                  | `60s`                               |
| `server.shutdown_timeout`   | `duration`             | Grace time to wait for active connections to close before shutdown                  | `5s`                                |
| `plugins`                   | `string[]`             | Comma-separated list of plugins to enable, see [Plugins](#plugins)                  | _none_                              |
| `temp.dir`                  | `string`               | Directory path for storing temporary files                                          | _none_ (= the OS temp dir)          |
| `eos.time`                  | `time`                 | End-of-service timestamp                                                            | _none_                              |
| `eos.message`               | `string`               | End-of-service message                                                              | _none_                              |
| `eos.url`                   | `string`               | End-of-service details URL                                                          | _none_                              |

Environment variable names must be prefixed by `RIPOSO_` and can be inferred by
capitalising the option name and replacing `.` with `_`. Examples:

```shell
# Set project.version
RIPOSO_PROJECT_VERSION=0.9.1

# Adjust server.shutdown_timeout
RIPOSO_SERVER_SHUTDOWN_TIMEOUT=30s

# Set permission defaults
RIPOSO_PERMISSION_DEFAULTS='{"bucket:create":[system.Authenticated], "other":[comma,separated,principals]}'
```

### ID Factories

ID factories are used to generate random, non-colliding, human-readable IDs for
records in the system. The two built-in methods are:

- `nanoid` - tiny, secure, URL-friendly, unique, see
  [Nano ID](https://github.com/ai/nanoid)
- `uuid` - v4 of the universally unique identifier, see
  [Wikipedia](<https://en.wikipedia.org/wiki/Universally_unique_identifier#Version_4_(random)>)

### Data Backends

Backends can be configured through URLs By default, your server comes with
support for two data backends:

- `:memory:` - purely in-memory, only use this for testing
- `postgres://` - PostgreSQL support, use this for production

Additional backends are available as [plugins](#plugins).

### Authentication

Authentication methods are available as plugins. By default only `basic` auth is
supported but additional methods are available as [plugins](#plugins).

### Default permissions

Default permissions will take precedence over those stored permanently in the
permission backend. For example:

```yaml
permission:
  defaults:
    # Allow all users to create new accounts
    account:create: [system.Everyone]
    # Allow authenticated users to create buckets
    bucket:create: [system.Authenticated]
    # Allow admin user to read all buckets
    bucket:read: [account:admin]
    # Grant admin user full access
    write: [account:admin]
```

The equivalent setting as env variable would be:

```shell
RIPOSO_PERMISSION_DEFAULT='{ account:create: [system.Everyone], bucket:create: [system.Authenticated], bucket:read: [account:admin], write: [account:admin] }'
```

### Plugins

Plugins can be loaded at runtime by referencing them via `RIPOSO_PLUGINS`
environment variable. To see a list of bundled plugins, please run
`riposo plugins`. By default, the following plugins are available:

- [accounts](https://github.com/riposo/accounts)
- [default-bucket](https://github.com/riposo/default-bucket)
- [flush](https://github.com/riposo/flush)

To build a custom release with additional plugins, please see instructions on
the [cmd](https://github.com/riposo/cmd) page. More plugins are available on
[GitHub](https://github.com/riposo?q=topic%3Aplugin).

## License

Copyright 2021 Black Square Media Ltd

Licensed under the Apache License, Version 2.0 (the "License"); you may not use
this material except in compliance with the License. You may obtain a copy of
the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the
specific language governing permissions and limitations under the License.
