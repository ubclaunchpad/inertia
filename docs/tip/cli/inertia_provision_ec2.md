## inertia provision ec2

[BETA] Provision a new Amazon EC2 instance

### Synopsis

[BETA] Provisions a new Amazon EC2 instance and sets it up for continuous deployment
with Inertia. 

Make sure you run this command with the '-p' flag to indicate what ports
your project uses - for example:

	inertia provision ec2 my_ec2_instance -p 8000

This ensures that your project ports are properly exposed and externally accessible.


```
inertia provision ec2 [name] [flags]
```

### Options

```
      --from-env              load ec2 credentials from environment - requires AWS_ACCESS_KEY_ID, AWS_ACCESS_KEY to be set
      --from-profile          load ec2 credentials from profile
  -h, --help                  help for ec2
      --profile.path string   path to aws profile credentials file (default "~/.aws/credentials")
      --profile.user string   user profile for aws credentials file (default "default")
  -t, --type string           ec2 instance type to instantiate (default "t2.micro")
  -u, --user string           ec2 instance user to execute commands as (default "ec2-user")
```

### Options inherited from parent commands

```
      --config string        specify relative path to Inertia project configuration (default "inertia.toml")
  -d, --daemon.port string   daemon port (default "4303")
  -p, --ports stringArray    ports your project uses
      --simple               disable colour and emoji output
```

### SEE ALSO

* [inertia provision](inertia_provision.md)	 - Provision a new remote host to deploy your project on

