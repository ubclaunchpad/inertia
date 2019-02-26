## inertia remote upgrade

Upgrade your remote configuration version to match the CLI

### Synopsis

Upgrade your remote configuration version to match the CLI and save it to global settings.

```
inertia remote upgrade [flags]
```

### Examples

```
inertia remote upgrade -r dev -r staging
```

### Options

```
  -h, --help                  help for upgrade
  -r, --remotes stringArray   specify which remotes to modify (default: all)
      --version string        specify Inertia daemon version to set (default "v0.5.2-22-gf564008")
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia configuration (default "inertia.toml")
```

### SEE ALSO

* [inertia remote](inertia_remote.md)	 - Configure the local settings for a remote host

