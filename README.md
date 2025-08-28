# xConfig

Lightweight, zero-dependency, and extendable configuration management.

xconfig is extremely light and extendable configuration management library with zero dependencies. Every aspect of configuration is provided through a _plugin_, which means you can have any combination of flags, environment variables, defaults, secret providers, Kubernetes Downward API, and what you want, and only what you want, through plugins.

xconfig takes the config schema as a struct decorated with tags, nesting is supported.

Supports all basic types, time.Duration, and any other type through `encoding.TextUnmarshaler` interface.
See the _[flat view](https://godoc.org/github.com/dv-net/xconfig/flat)_ package for details.

## Examples

### Load config

```go
func loadConfig(configPaths []string) (*config.Config, error) {
  conf := new(config.Config)

  files := xconfig.Files{}
  for _, path := range configPaths {
    files.Add(path, yaml.Unmarshal, false) // json.Unmarshal and any other
  }

  _, err := xconfig.Load(conf,
    xconfig.WithFiles(files),
    xconfig.WithPlugins(
      validate.New(func(a any) error {
        return validator.New().Struct(a)
      }),
    ),
  )
  if err != nil {
    return nil, fmt.Errorf("failed to load configuration: %w", err)
  }

  return conf, nil
}
```

### Generate default envs

```go
func generateDefaultEnvs() error {
  conf := new(config.Config)

  // generate markdown
  markdown, err := xconfig.GenerateMarkdown(conf)
  if err != nil {
    return fmt.Errorf("failed to generate markdown: %w", err)
  }

  if err := os.WriteFile("ENVS.md", []byte(markdown), os.ModePerm); err != nil {
    return err
  }

  // generate yaml
  buf := bytes.NewBuffer(nil)
  enc := yaml.NewEncoder(buf, yaml.Indent(2))
  defer enc.Close()

  if err := enc.Encode(conf); err != nil {
    return fmt.Errorf("failed to encode yaml: %w", err)
  }

  if err := os.WriteFile("config.template.yaml", buf.Bytes(), os.ModePerm); err != nil {
    return fmt.Errorf("failed to write file: %w", err)
  }

  return nil
}
```

## Available plugins

- defaults
- custom defaults
- env
- flag
- validate

## TODO

- [ ] Generate flags
- [ ] AllowUnknownFields
- [ ] AllowDuplicates
- [ ] AllFieldRequired
- [ ] MergeFiles
