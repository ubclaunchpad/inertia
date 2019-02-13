## inertia [remote_name] init

Initialize remote host for deployment

### Synopsis

Initializes this remote host for deployment.

This command sets up your remote host and brings an Inertia daemon online on your remote.

Upon successful setup, you will be provided with:
	- a deploy key
	- a webhook URL

The deploy key is required for the daemon to access your repository, and the
webhook URL enables continuous deployment as your repository is updated.

```
inertia [remote_name] init [flags]
```

### Options

```
  -h, --help   help for init
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia configuration (default "inertia.toml")
  -s, --short           don't stream output from command
      --verify-ssl      verify SSL communications - requires a signed SSL certificate
```

### SEE ALSO

* [inertia [remote_name]](inertia_[remote_name].md)	 - Configure deployment to [remote_name]

