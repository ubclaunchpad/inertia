## inertia ${remote_name} user login

Authenticate with the remote as a user

### Synopsis

Retreives an access token from the remote using your credentials.
	
If this remote was previously authenticated against as a user, then the user
argument is optional.

```
inertia ${remote_name} user login [user] [flags]
```

### Options

```
  -h, --help          help for login
      --totp string   auth code or backup code for 2FA
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia project configuration (default "inertia.toml")
      --debug           enable debug output from Inertia client
  -s, --short           don't stream output from command
      --simple          disable colour and emoji output
```

### SEE ALSO

* [inertia ${remote_name} user](inertia_${remote_name}_user.md)	 - Configure user access to Inertia Web

