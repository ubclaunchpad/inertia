## inertia remote upgrade

Upgrade your remote configuration version to match the CLI

### Synopsis

Upgrade your remote configuration version to match the CLI and save it to global settings.

```
inertia remote upgrade [flags]
```

### Examples

```
inertia remote upgrade dev staging
```

### Options

```
      --all              upgrade all remotes
  -h, --help             help for upgrade
      --version string   specify Inertia daemon version to set (default "v0.6.0-preview2-24-g44442aa")
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia project configuration (default "inertia.toml")
      --simple          disable colour and emoji output
```

### SEE ALSO

* [inertia remote](inertia_remote.md)	 - Configure the local settings for a remote host

