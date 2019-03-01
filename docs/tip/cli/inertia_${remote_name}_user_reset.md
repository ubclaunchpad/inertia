## inertia ${remote_name} user reset

Reset user database on your remote

### Synopsis

Removes all users credentials on your remote. All configured user
will no longer be able to log in and view or configure the deployment
remotely.

```
inertia ${remote_name} user reset [flags]
```

### Options

```
  -h, --help   help for reset
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia configuration (default "inertia.toml")
      --debug           enable debug out from Inertia client
  -s, --short           don't stream out from command
      --simple          disable colour output
```

### SEE ALSO

* [inertia ${remote_name} user](inertia_${remote_name}_user.md)	 - Configure user access to Inertia Web

