## inertia project profile apply

Apply a project configuration profile to a remote

### Synopsis

Applies a project configuration profile to an existing remote. The applied
profile will be used whenever you run 'inertia ${remote_name} up' on the target
remote.

By default, the profile called 'default' will be used.

```
inertia project profile apply [profile] [remote] [flags]
```

### Options

```
  -h, --help   help for apply
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia project configuration (default "inertia.toml")
      --simple          disable colour and emoji output
```

### SEE ALSO

* [inertia project profile](inertia_project_profile.md)	 - Manage project profile configurations

