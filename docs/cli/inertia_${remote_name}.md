## inertia ${remote_name}

Configure deployment to ${remote_name}

### Synopsis

Manages deployment on specified remote.

Requires:
1. an Inertia daemon running on your remote - use 'inertia [remote] init' to get it running.
2. a deploy key to be registered within your remote repository for the daemon to use.

Continuous deployment requires the daemon's webhook address to be registered in your remote repository.

If the SSH key for your remote requires a passphrase, it can be provided via 'IDENTITY_PASSPHRASE'.

Run 'inertia [remote] init' to gather this information.

### Options

```
      --debug   enable debug output from Inertia client
  -h, --help    help for ${remote_name}
  -s, --short   don't stream output from command
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia project configuration (default "inertia.toml")
      --simple          disable colour and emoji output
```

### SEE ALSO

* [inertia](inertia.md)	 - Effortless, self-hosted continuous deployment for small teams and projects
* [inertia ${remote_name} down](inertia_${remote_name}_down.md)	 - Bring project offline on remote
* [inertia ${remote_name} env](inertia_${remote_name}_env.md)	 - Manage environment variables on your remote
* [inertia ${remote_name} init](inertia_${remote_name}_init.md)	 - Initialize remote host for deployment
* [inertia ${remote_name} logs](inertia_${remote_name}_logs.md)	 - Access logs of containers on your remote host
* [inertia ${remote_name} prune](inertia_${remote_name}_prune.md)	 - Prune Docker assets and images on your remote
* [inertia ${remote_name} send](inertia_${remote_name}_send.md)	 - Send a file to your Inertia deployment
* [inertia ${remote_name} ssh](inertia_${remote_name}_ssh.md)	 - Start an interactive SSH session
* [inertia ${remote_name} status](inertia_${remote_name}_status.md)	 - Print the status of the deployment on this remote
* [inertia ${remote_name} token](inertia_${remote_name}_token.md)	 - Generate tokens associated with permission levels for admin to share.
* [inertia ${remote_name} uninstall](inertia_${remote_name}_uninstall.md)	 - Shut down Inertia and remove Inertia assets from remote host
* [inertia ${remote_name} up](inertia_${remote_name}_up.md)	 - Bring project online on remote
* [inertia ${remote_name} upgrade](inertia_${remote_name}_upgrade.md)	 - Upgrade Inertia daemon to match the CLI.
* [inertia ${remote_name} user](inertia_${remote_name}_user.md)	 - Configure user access to Inertia Web

