# proxyutil

Proxyutil is a small utility for proxying requests to a server.

It is useful if you only have/want a single port open to multiple routes;

## Usage

```bash
proxyutil '/:http://localhost:3000' '/api/v1:http://localhost:8070' '/api/v2:http://localhost:8080'
```

Will allow you to proxy requests to `http://localhost:3000` and `http://localhost:8070/api/v1` and `http://localhost:8080/api/v2` respectively.

You can also use a config file to specify the routes:

```proxyutil
/:http://localhost:3000
/api/v1:http://localhost:8070
/api/v2:http://localhost:8080
```

By default, file is looked for in `$CWD/.proxies` but can be configured with `--config` or `$PROXYUTIL_CONFIG`.
