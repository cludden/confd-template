# confd-template
a `cli` for generating [confd](https://github.com/kelseyhightower/confd) templates using a populated KV backend

```
$ confd-template
a cli for generating confd templates using a populated KV backend

Usage:
  confd-template [flags]

Flags:
      --backend string     configuration backend (default "ssm")
      --delimiter string   key delimiter (default "/")
      --filter string      optional regex key filter
      --format string      template format (default "yaml")
  -h, --help               help for confd-template
      --level string       log level (default "info")
      --out string         output template path
      --prefix string      key prefix to scan (default "/")
```

## Installation

### Binary Download

You can download the latest release from [GitHub](https://github.com/cludden/confd-template/releases)

```
$ wget https://github.com/cludden/confd-template/releases/download/v{version}/confd-template-{version}-{os}-amd64
```

Ensure the binary is in your path and is executable (*these commands may require `sudo`*)
```
$ mv confd-template-{version}-{os}-amd64 /usr/local/bin/confd-template
$ chmod +x /usr/local/bin/confd-template
```

## Backends
- [x] ssm

## Formatters
- [x] yaml
- [ ] json

## License
Copyright (c) 2018 Chris Ludden

Licensed under the [MIT License](LICENSE.md)