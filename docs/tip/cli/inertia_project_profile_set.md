## inertia project profile set

Configure project profiles

### Synopsis

Configures project profiles - if the given profile does not exist,
a new one is created, otherwise the existing one is overwritten.

Provide profile values via the available flags.

```
inertia project profile set [profile] [flags]
```

### Examples

```
inertia project profile set my_profile --build.type dockerfile --build.file Dockerfile.dev
```

### Options

```
      --branch string       branch for profile (default: current branch)
      --build.file string   relative path to build config file (e.g. 'Dockerfile')
      --build.type string   build type for profile
  -h, --help                help for set
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia configuration (default "inertia.toml")
      --simple          disable colour output
```

### SEE ALSO

* [inertia project profile](inertia_project_profile.md)	 - Manage project profile configurations

