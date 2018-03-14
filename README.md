# confd-template
a `cli` for generating [confd](https://github.com/kelseyhightower/confd) templates using a populated KV backend

```
$ confd-template --help
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

## Getting Started
Basic cli usage is shown below. This will use the default backend (ssm) and render format (yaml) to output a confd template containing all of the keys at prefix "/secrets/production-us-east-1" that begin with "/foo/" or "/bar/".
```
$ confd-template --level debug --out config.yml.tmpl --prefix /secrets/production-us-east-1 --filter "^/(foo|bar)/*"
```

## Installation

You can download the latest release from [GitHub](https://github.com/cludden/confd-template/releases)

```
$ wget https://github.com/cludden/confd-template/releases/download/v{version}/confd-template-{version}-{os}-amd64
```

Ensure the binary is in your path and is executable (*these commands may require `sudo`*)
```
$ mv confd-template-{version}-{os}-amd64 /usr/local/bin/confd-template
$ chmod +x /usr/local/bin/confd-template
```

## Todo
**General:**
- [ ] test test test

**Backends:**
- [x] ssm

**Formatters:**
- [x] yaml
- [ ] json

## Contributing
1. [Fork it](https://github.com/cludden/confd-template/fork)
1. Create your feature branch (`git checkout -b my-new-feature`)
1. Commit your changes using [conventional changelog standards](https://github.com/bcoe/conventional-changelog-standard/blob/master/convention.md) (`git commit -am 'feat: adds my new feature'`)
1. Push to the branch (`git push origin my-new-feature`)
1. Ensure lint/tests are all passing
1. Create new Pull Request

## License
Copyright (c) 2018 Chris Ludden

Licensed under the [MIT License](LICENSE.md)