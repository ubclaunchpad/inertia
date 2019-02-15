## inertia ${remote_name} up

Bring project online on remote

### Synopsis

Builds and deploy your project on your remote.

This requires an Inertia daemon to be active on your remote - do this by running 'inertia [remote] init'

```
inertia ${remote_name} up [flags]
```

### Options

```
  -h, --help          help for up
      --type string   override configured build method for your project
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia configuration (default "inertia.toml")
  -s, --short           don't stream output from command
      --verify-ssl      verify SSL communications - requires a signed SSL certificate
```

### SEE ALSO

* [inertia ${remote_name}](inertia_${remote_name}.md)	 - Configure deployment to ${remote_name}

