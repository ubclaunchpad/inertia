## inertia project profile configure

Configure project profiles

### Synopsis

Configures project profiles - if the given profile does not exist,
a new one is created, otherwise the existing one is overwritten.

Provide profile values via the available flags.

```
inertia project profile configure [profile] [flags]
```

### Examples

```
inertia project profile configure my_profile --build.type dockerfile --build.file Dockerfile.dev
```

### Options

```
      --branch string       branch for profile (default: current branch)
      --build.file string   relative path to build config file (e.g. 'Dockerfile')
      --build.type string   build type for profile
  -h, --help                help for configure
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia project configuration (default "inertia.toml")
      --simple          disable colour and emoji output
```

### SEE ALSO

* [inertia project profile](inertia_project_profile.md)	 - Manage project profile configurations

