## inertia remote login

Log in to a remote as an existing user.

### Synopsis

Log in as an existing user to access a remote.

```
inertia remote login [remote] [user] [flags]
```

### Examples

```
inertia remote login staging my_user --ip 1.2.3.4
```

### Options

```
      --daemon.port string   remote daemon port (default "4303")
  -h, --help                 help for login
      --ip string            IP address of remote
      --totp string          auth code or backup code for 2FA
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia project configuration (default "inertia.toml")
      --simple          disable colour and emoji output
```

### SEE ALSO

* [inertia remote](inertia_remote.md)	 - Configure the local settings for a remote host

