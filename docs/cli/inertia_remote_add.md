## inertia remote add

Add a reference to a remote VPS instance

### Synopsis

Adds a reference to a remote VPS instance. Requires information about the VPS
including IP address, user and a PEM file. The provided name will be used in other
Inertia commands.

```
inertia remote add [remote] [flags]
```

### Options

```
  -h, --help              help for add
  -p, --port string       remote daemon port (default "4303")
  -s, --ssh.port string   remote SSH port (default "22")
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia configuration (default "inertia.toml")
```

### SEE ALSO

* [inertia remote](inertia_remote.md)	 - Configure the local settings for a remote host

