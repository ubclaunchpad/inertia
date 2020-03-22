## inertia remote add

Add a reference to a remote VPS instance

### Synopsis

Adds a reference to a remote VPS instance. Requires information about the VPS
including IP address, user and a identity file. The provided name will be used in other
Inertia commands.

```
inertia remote add [remote] [flags]
```

### Examples

```
inertia remote add staging --daemon.gen-secret --ip 1.2.3.4
```

### Options

```
      --daemon.gen-secret    toggle webhook secret generation (default true)
      --daemon.port string   remote daemon port (default "4303")
  -h, --help                 help for add
      --ip string            IP address of remote
      --ssh.key string       path to SSH key for remote
      --ssh.port string      remote SSH port (default "22")
      --ssh.user string      user to use when accessing remote over SSH
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia project configuration (default "inertia.toml")
      --simple          disable colour and emoji output
```

### SEE ALSO

* [inertia remote](inertia_remote.md)	 - Configure the local settings for a remote host

