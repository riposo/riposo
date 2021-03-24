# Riposo Server

Riposo is a JSON document store written in [Go](https://golang.org), based on the design of [Kinto](https://docs.kinto-storage.org/en/latest/index.html). It implements [v1.22](https://docs.kinto-storage.org/en/latest/api/index.html) of the Kinto API and aims to be fully compatible with the client libraries [kinto-http.js](https://github.com/Kinto/kinto-http.js), [kinto-http.py](https://github.com/Kinto/kinto-http.py) as well as higher-level, offline-first abstraction [kinto.js](https://github.com/Kinto/kinto.js).

## Getting Started

The easiest way to deploy Riposo is via Docker.

To launch a test instance (on port 8888):

```shell
docker run --rm -p 8888:8888 riposo/riposo server
```

You can use environment variables to configure all aspects of your instance, including persistence. For more information, please see the [Configuration](#configuration) section.

```shell
docker run --rm \
    -e RIPOSO_STORAGE_URL=postgres://postgres:postgres@database/riposo?timezone=UTC \
    -e RIPOSO_PERMISSION_URL=postgres://postgres:postgres@database/riposo?timezone=UTC \
    -e RIPOSO_CACHE_URL=postgres://postgres:postgres@database/riposo?timezone=UTC \
    riposo/riposo:0.1.0 server
```

Plugins can also be loaded dynamically by referencing the plugin ID. In order to enable a plugin, the server binary must have been built with it. For more information, please see the [Plugins](#plugins) section.

```shell
# list plugins
docker run --rm riposo/riposo plugins

# run server with plugins
docker run --rm \
    -v $(pwd):/data
    -e RIPOSO_PLUGINS=accounts,default-bucket \
    -e RIPOSO_BUCKET_CREATE_PRINCIPALS=system.Authenticated \
    -e RIPOSO_ACCOUNT_CREATE_PRINCIPALS=system.Everyone \
    riposo/riposo server
```

## Configuration

The following environment variable can be used to configure your server instance:

| Variable                           | Description                                                                 | Default                             |
| ---------------------------------- | --------------------------------------------------------------------------- | ----------------------------------- |
| `RIPOSO_PROJECT_NAME`              | Project name, as reported by `GET /v1/`                                     | `riposo`                            |
| `RIPOSO_PROJECT_VERSION`           | Project version, as reported by `GET /v1/`                                  | _none_                              |
| `RIPOSO_PROJECT_DOCS`              | Project documentation URL                                                   | `https://github.com/riposo/riposo/` |
| `RIPOSO_ID_FACTORY`                | ID generator, see [ID Factories](#id-factories)                             | `nanoid`                            |
| `RIPOSO_STORAGE_URL`               | Storage backend URL, see [Data Backends](#data-backends)                    | `:memory:`                          |
| `RIPOSO_PERMISSION_URL`            | Permission backend URL                                                      | `:memory:`                          |
| `RIPOSO_CACHE_URL`                 | Cache backend URL                                                           | `:memory:`                          |
| `RIPOSO_BATCH_MAX_REQUESTS`        | Maximum permitted number of requests per batch                              | `25`                                |
| `RIPOSO_AUTH_METHODS`              | Comma-separated list of auth methods, see [Authentication](#authentication) | `basic`                             |
| `RIPOSO_AUTH_HASH`                 | Hash method used for password hashing, `argon2id` or `bcrypt`               | `argon2id`                          |
| `RIPOSO_CORS_ORIGINS`              | Permitted CORS origins                                                      | `*`                                 |
| `RIPOSO_CORS_MAX_AGE`              | Indicates how long the results of a preflight CORS request can be cached    | `1h`                                |
| `RIPOSO_PAGINATION_TOKEN_VALIDITY` | Pagination token TTL                                                        | `10m`                               |
| `RIPOSO_PAGINATION_MAX_LIMIT`      | Maximum number of records that can be requested per page                    | `10000`                             |
| `RIPOSO_SERVER_ADDR`               | Server address to listen for requests                                       | `:8888`                             |
| `RIPOSO_SERVER_READ_TIMEOUT`       | Read timeout for server requests                                            | `60s`                               |
| `RIPOSO_SERVER_WRITE_TIMEOUT`      | Write timeout for server responses                                          | `60s`                               |
| `RIPOSO_SERVER_SHUTDOWN_TIMEOUT`   | Grace time to wait for active connections to close before shutdown          | `5s`                                |
| `RIPOSO_PLUGINS`                   | Comma-separated list of plugins to enable, see [Plugins](#plugins)          | _none_                              |
| `RIPOSO_TEMP_DIR`                  | Directory path for storing temporary files                                  | _none_ (= the OS temp dir)          |
| `RIPOSO_EOS_TIME`                  | End-of-service timestamp                                                    | _none_                              |
| `RIPOSO_EOS_MESSAGE`               | End-of-service message                                                      | _none_                              |
| `RIPOSO_EOS_URL`                   | End-of-service details URL                                                  | _none_                              |

Additionally, you can use extra variables to grant permissions globally. These will take precedence over those stored permanently in the permission backend.
The format of these special variables is `RIPOSO_<permission>_PRINCIPALS=comma,separated,principals`. Examples:

```shell
# Allow authenticated users to create buckets
RIPOSO_BUCKET_CREATE_PRINCIPALS=system.Authenticated

# Allow all users to create new accounts
RIPOSO_ACCOUNT_CREATE_PRINCIPALS=system.Everyone

# Allow admin user to read all buckets
RIPOSO_BUCKET_READ_PRINCIPALS=account:admin

# Grant admin user full access
RIPOSO_WRITE_PRINCIPALS=account:admin
```

### ID Factories

ID factories are used to generate random, non-colliding, human-readable IDs for records in the system. The two built-in methods
are:

* `nanoid` - tiny, secure, URL-friendly, unique, see [Nano ID](https://github.com/ai/nanoid)
* `uuid` - v4 of the universally unique identifier, see [Wikipedia](https://en.wikipedia.org/wiki/Universally_unique_identifier#Version_4_(random))

### Data Backends

Backends can be configured through URLs By default, your server comes with support for two data backends:

* `:memory:` - purely in-memory, only use this for testing
* `postgres://` - PostgreSQL support, use this for production

Additional backends are available as [plugins](#plugins).

### Authentication

Authentication methods are available as plugins. By default only `basic` auth is supported but additional methods are available as [plugins](#plugins).

### Plugins

Plugins can be loaded at runtime by referencing them via `RIPOSO_PLUGINS` environment variable. To see a list of bundled plugins, please run `riposo plugins`. By default, the following plugins are available:

* [accounts](https://github.com/riposo/accounts)
* [default-bucket](https://github.com/riposo/default-bucket)
* [flush](https://github.com/riposo/flush)

To build a custom release with additional plugins, please see instructions on the [Composer](https://github.com/riposo/composer) page. More plugins are available on [GitHub](https://github.com/riposo?q=topic%3Aplugin).

## License

Copyright 2021 Black Square Media Ltd

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this material except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
