## inertia ${remote_name} env rm

Remove an environment variable from your remote

### Synopsis

Removes the specified environment variable from deployed containers
and persistent environment storage.

```
inertia ${remote_name} env rm [name] [flags]
```

### Options

```
  -h, --help   help for rm
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia project configuration (default "inertia.toml")
      --debug           enable debug output from Inertia client
  -s, --short           don't stream output from command
      --simple          disable colour and emoji output
```

### SEE ALSO

* [inertia ${remote_name} env](inertia_${remote_name}_env.md)	 - Manage environment variables on your remote

