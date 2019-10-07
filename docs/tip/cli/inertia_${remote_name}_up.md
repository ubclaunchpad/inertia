## inertia ${remote_name} up

Bring project online on remote

### Synopsis

Builds and deploy your project on your remote using your project's
default profile, or a profile you have applied using 'inertia project profile apply'.

This requires an Inertia daemon to be active on your remote - do this by running
'inertia [remote] init'.

```
inertia ${remote_name} up [flags]
```

### Options

```
  -h, --help             help for up
  -p, --profile string   specify a profile to deploy
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia project configuration (default "inertia.toml")
      --debug           enable debug output from Inertia client
  -s, --short           don't stream output from command
      --simple          disable colour and emoji output
```

### SEE ALSO

* [inertia ${remote_name}](inertia_${remote_name}.md)	 - Configure deployment to ${remote_name}

