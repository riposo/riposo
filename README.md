# Riposo Server

## Configuration

The following environment variable can be used to configure your server instance:

| Variable                           | Description                                                                           | Default                             |
| ---------------------------------- | ------------------------------------------------------------------------------------- | ----------------------------------- |
| `RIPOSO_PROJECT_NAME`              | Project name, as reported by `GET /v1/`                                               | `riposo`                            |
| `RIPOSO_PROJECT_VERSION`           | Project version, as reported by `GET /v1/`                                            | _none_                              |
| `RIPOSO_PROJECT_DOCS`              | Project documentation URL                                                             | `https://github.com/riposo/riposo/` |
| `RIPOSO_ID_FACTORY`                | ID generator                                         (ID Factories)(#id-factories)    | `nanoid`                            |
| `RIPOSO_STORAGE_URL`               | Storage backend URL, see (Data Backends)(#data-backends)                              | `:memory:`                          |
| `RIPOSO_PERMISSION_URL`            | Permission backend URL                                                                | `:memory:`                          |
| `RIPOSO_CACHE_URL`                 | Cache backend URL                                                                     | `:memory:`                          |
| `RIPOSO_BATCH_MAX_REQUESTS`        | Maximum permitted number of requests per batch                                        | `25`                                |
| `RIPOSO_AUTH_METHODS`              | Comma-separated list of permitted auth methods, see (Authentication)(#authentication) | `basic`                             |
| `RIPOSO_AUTH_HASH`                 | Hash method used for password hashing                                                 | `argon2id`                          |
| `RIPOSO_CORS_ORIGINS`              | Permitted CORS origins                                                                | `*`                                 |
| `RIPOSO_CORS_MAX_AGE`              | Indicates how long the results of a preflight CORS request can be cached              | `1h`                                |
| `RIPOSO_PAGINATION_TOKEN_VALIDITY` | Pagination token TTL                                                                  | `10m`                               |
| `RIPOSO_PAGINATION_MAX_LIMIT`      | Maximum number of records that can be requested per page                              | `10000`                             |
| `RIPOSO_SERVER_ADDR`               | Server address to listen for requests                                                 | `:8888`                             |
| `RIPOSO_SERVER_READ_TIMEOUT`       | Read timeout for server requests                                                      | `60s`                               |
| `RIPOSO_SERVER_WRITE_TIMEOUT`      | Write timeout for server responses                                                    | `60s`                               |
| `RIPOSO_SERVER_SHUTDOWN_TIMEOUT`   | Grace time to wait for active connections to close before shutdown                    | `5s`                                |
| `RIPOSO_PLUGIN_DIR`                | Local plugin directory path(s), see (Plugins)(#plugins)                               | `./plugins`                         |
| `RIPOSO_PLUGIN_URLS`               | Commas-separated list of URLs to fetch plugins from                                   | _none_                              |
| `RIPOSO_TEMP_DIR`                  | Directory path for storing temporary files                                            | _none_ (= the OS temp dir)          |
| `RIPOSO_EOS_TIME`                  | End-of-service timestamp                                                              | _none_                              |
| `RIPOSO_EOS_MESSAGE`               | End-of-service message                                                                | _none_                              |
| `RIPOSO_EOS_URL`                   | End-of-service details URL                                                            | _none_                              |

Additionally, you can use extra variables to grant permissions globally. These will take precedence over those stored permanently in the permission backend.
The format of these special variables is `RIPOSO_<permission>_PRINCIPALS=comma,separated,principals`. Examples:

```
## Allow authenticated users to create buckets
RIPOSO_BUCKET_CREATE_PRINCIPALS=system.Authenticated

## Allow all users to create new accounts
RIPOSO_ACCOUNT_CREATE_PRINCIPALS=system.Everyone

## Allow admin user to read all buckets
RIPOSO_BUCKET_READ_PRINCIPALS=account:admin

## Grant admin user full access
RIPOSO_WRITE_PRINCIPALS=account:admin
```

### ID Factories

ID factories are used to generate random, non-colliding, human-readable IDs for records in the system. The two built-in methods
are:

* `nanoid` - tiny, secure, URL-friendly, unique, see (Nano ID)[https://github.com/ai/nanoid]
* `uuid` - v4 of the universally unique identifier, see (Wikipedia)[https://en.wikipedia.org/wiki/Universally_unique_identifier#Version_4_(random)]

### Data Backends

Backends can be configured through URLs By default, your server comes with support for two data backends:

* `:memory:` - purely in-memory, only use this for testing
* `postgres://` - PostgreSQL support, use this for production

Additional backends are available as plugins.

### Authentication

Authentication methods are available as plugins. By default only `basic` auth is supported which requires the [accounts](https://github.com/riposo/accounts)
plugin.

Additionally, two built-in methods for password hashing can be configured: `argon2id` (default) and `bcrypt`.

### Plugins

Plugins can be loaded at runtime from a list of local directory paths (comma-separated). Additionally, plugins can also be automatically fetched
from remote URLs. Downloaded plugins are stored in the `${RIPOSO_TEMP_DIR}/riposo-plugins` directory.
