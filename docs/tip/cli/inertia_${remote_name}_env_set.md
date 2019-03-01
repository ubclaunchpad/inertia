## inertia ${remote_name} env set

Set an environment variable on your remote

### Synopsis

Sets a persistent environment variable on your remote. Set environment
variables are applied to all deployed containers.

```
inertia ${remote_name} env set [name] [value] [flags]
```

### Options

```
  -e, --encrypt   encrypt variable when stored
  -h, --help      help for set
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia configuration (default "inertia.toml")
      --debug           enable debug out from Inertia client
  -s, --short           don't stream out from command
      --simple          disable colour output
```

### SEE ALSO

* [inertia ${remote_name} env](inertia_${remote_name}_env.md)	 - Manage environment variables on your remote

