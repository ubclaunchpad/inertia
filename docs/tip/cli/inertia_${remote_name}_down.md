## inertia ${remote_name} down

Bring project offline on remote

### Synopsis

Stops your project on your remote. This will kill all active project containers on your remote.
	
Requires project to be online - do this by running 'inertia [remote] up

```
inertia ${remote_name} down [flags]
```

### Options

```
  -h, --help   help for down
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia configuration (default "inertia.toml")
  -s, --short           don't stream output from command
```

### SEE ALSO

* [inertia ${remote_name}](inertia_${remote_name}.md)	 - Configure deployment to ${remote_name}

