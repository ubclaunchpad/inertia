## inertia ${remote_name} status

Print the status of the deployment on this remote

### Synopsis

Prints the status of the deployment on this remote.

Requires the Inertia daemon to be active on your remote - do this by running 'inertia [remote] up'

```
inertia ${remote_name} status [flags]
```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia configuration (default "inertia.toml")
      --debug           enable debug out from Inertia client
  -s, --short           don't stream out from command
      --simple          disable colour output
```

### SEE ALSO

* [inertia ${remote_name}](inertia_${remote_name}.md)	 - Configure deployment to ${remote_name}

