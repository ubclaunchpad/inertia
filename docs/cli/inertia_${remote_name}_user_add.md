## inertia ${remote_name} user add

Create a user with access to this remote's Inertia daemon

### Synopsis

Creates a user with access to this remote's Inertia daemon.

This user will be able to log in and view or configure the deployment
from the Inertia CLI (using 'inertia [remote] user login').

Use the --admin flag to create an admin user.

```
inertia ${remote_name} user add [user] [flags]
```

### Options

```
      --admin   create a user with administrator permissions
  -h, --help    help for add
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

