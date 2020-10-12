OAuth2 Authenticating Proxy ![Publish Docker image](https://github.com/matchory/oauth2-authenticating-proxy/workflows/Publish%20Docker%20image/badge.svg)
===========================
> Simple proxy application to add an OAuth2 access token to any request passing through

This application will take any incoming HTTP request and attach a valid OAuth2 token in a standard `Authorization`
header as `Bearer {{TOKEN}}`. This comes in handy if you have a trusted application making requests that should be
authorized automatically.  
To do so, the proxy automatically fetches an access token from the authorization server you configured. It will try to
refresh it automatically and reuse an existing token, if possible.

**Attention: Using this proxy can be dangerous. Make sure you strictly limit access to it!**

Installation
------------
We strongly recommend using the proxy in a Docker stack, e.g. `docker-compose`:

```yaml
version: 3.7
services:
  proxy:
    image: ghcr.io/matchory/oauth2-authenticating-proxy:latest
    expose:
      - 8080
    volumes:
      - "./proxy.yaml:/proxy/config.yaml:ro"
```

Configuration
-------------
The proxy supports passing configuration from environment variables or from a YAML configuration file. Environment
variables will always take precedence over values found in the configuration file.  
The following options are supported:

| Option            | Environment variable  | Default | Description                                                 |
|:------------------|:----------------------|:-------:|:------------------------------------------------------------|
| `listen_port`     | `LISTEN_PORT`         | `8080`  | Network port to listen on.                                  |
| `skip_tls_verify` | `SKIP_TLS_VERIFY`     | `false` | Whether to skip TLS certificate validation.                 |
| `client_id`       | `CLIENT_ID`           |    -    | OAuth2 client ID to authenticate with.                      |
| `client_secret`   | `CLIENT_SECRET`       |    -    | OAuth2 client secret to authenticate with.                  |
| `token_endpoint`  | `TOKEN_ENDPOINT`      |    -    | Fully qualified URL of your OAuth2 token endpoint.          |
| `scopes`          | `SCOPES`              |    -    | List of scopes to request for the token.                    |
| `upstream_scheme` | `UPSTREAM_SCHEME`     | `https` | URL scheme to use for _upstream_ connections.               |
| `upstream_host`   | `UPSTREAM_HOST`       |    -    | Host to forward requests to. [Optional](#upstream-hosts).   |
| `allowed_hosts`   | `ALLOWED_HOSTS`       |    -    | List of allowed hosts to forward to. Optional.              |

Usage
-----
After spinning up the image with correct configuration, you should be able to send HTTP requests _without_ an
`Authorization` header to the proxy and see requests _with_ the header, and a valid token arrive at your back service.

### Upstream hosts

When proxying a request, it needs to be sent to the proxy host instead of the actual, intended host. To make this
possible, you'll need some way to tell the proxy server where to send the modified request to. The OAuth2 proxy provides
you with three different ways to resolve the target host:

1. **Set the `upstream_host` configuration directive**  
   If all you ever need to do is send requests to a single upstream, you can set the hostname (without a protocol) in
   your configuration file (or using the `UPSTREAM_HOST` environment variable), and all requests will be forwarded to
   that host.
2. **Set the `Forward` request header**  
   To dynamically set the forward host, you can set the `Forward` header on your requests. This will even take
   precedence over the configured upstream host from variant 1. To make it harder to shoot yourself in the foot, you
   can (and should!) configure the `allowed_hosts` setting with all hosts you explicitly want to talk to.
3. **Set the `Host` request header independently**  
   Depending on the type of library you use, you have the possibility to set the `Host` header independently of the
   request URI: The proxy uses the value of the host header as the fallback value, if none of the other two methods
   resolved a hostname.
