## inertia ${remote_name} env

Manage environment variables on your remote

### Synopsis

Manages environment variables on your remote through Inertia. 
			
Configured variables can be encrypted or stored in plain text. They are applied
as follows:

- for docker-compose projects, variables are set for the docker-compose process
- for Dockerfile projects, variables are set in the deployed container


### Options

```
  -h, --help   help for env
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
* [inertia ${remote_name} env ls](inertia_${remote_name}_env_ls.md)	 - List currently set and saved environment variables
* [inertia ${remote_name} env rm](inertia_${remote_name}_env_rm.md)	 - Remove an environment variable from your remote
* [inertia ${remote_name} env set](inertia_${remote_name}_env_set.md)	 - Set an environment variable on your remote

