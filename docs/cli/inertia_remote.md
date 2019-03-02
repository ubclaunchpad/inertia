## inertia remote

Configure the local settings for a remote host

### Synopsis

Configures local settings for a remote host - add, remove, and list configured
Inertia remotes.

Requires Inertia to be set up via 'inertia init'.

For example:
inertia init
inertia remote add gcloud
inertia gcloud init        # set up Inertia
inertia gcloud status      # check on status of Inertia daemon


### Options

```
  -h, --help   help for remote
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia configuration (default "inertia.toml")
```

### SEE ALSO

* [inertia](inertia.md)	 - Effortless, self-hosted continuous deployment for small teams and projects
* [inertia remote add](inertia_remote_add.md)	 - Add a reference to a remote VPS instance
* [inertia remote ls](inertia_remote_ls.md)	 - List currently configured remotes
* [inertia remote rm](inertia_remote_rm.md)	 - Remove a configured remote
* [inertia remote set](inertia_remote_set.md)	 - Update details about remote
* [inertia remote show](inertia_remote_show.md)	 - Show details about a remote
* [inertia remote upgrade](inertia_remote_upgrade.md)	 - Upgrade your remote configuration version to match the CLI

