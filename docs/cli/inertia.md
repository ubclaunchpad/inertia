## inertia

Effortless, self-hosted continuous deployment for small teams and projects

### Synopsis

Inertia is an effortless, self-hosted continuous deployment platform 🚀 

Initialization involves preparing a server to run an application, then
activating a daemon which will continuously update the production server
with new releases as they become available in the project's repository.

Once you have set up a remote with 'inertia remote add [remote]', use 
'inertia [remote] --help' to see what you can do with your remote. To list
available remotes, use 'inertia remote ls'.

Global inertia configuration is stored in '~/.inertia'.

💻  Repository:    https://github.com/ubclaunchpad/inertia/
🎫  Issue tracker: https://github.com/ubclaunchpad/inertia/issues
📚  Documentation: https://inertia.ubclaunchpad.com

### Options

```
      --config string   specify relative path to Inertia project configuration (default "inertia.toml")
  -h, --help            help for inertia
      --simple          disable colour and emoji output
```

### SEE ALSO

* [inertia ${remote_name}](inertia_${remote_name}.md)	 - Configure deployment to ${remote_name}
* [inertia init](inertia_init.md)	 - Initialize an Inertia project in this repository
* [inertia project](inertia_project.md)	 - Update and configure Inertia project settings
* [inertia provision](inertia_provision.md)	 - Provision a new remote host to deploy your project on
* [inertia remote](inertia_remote.md)	 - Configure the local settings for a remote host

