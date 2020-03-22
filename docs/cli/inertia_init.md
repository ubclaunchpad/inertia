## inertia init

Initialize an Inertia project in this repository

### Synopsis

Initializes an Inertia project in this GitHub repository. You can
provide an argument as the name of your project, otherwise the name of your
current directory will be used.

There must be a local git repository in order for initialization
to succeed, unless you use the '--global' flag to initialize only
the Inertia global configuration.

See https://inertia.ubclaunchpad.com/#project-configuration for more details.

```
inertia init [flags]
```

### Examples

```
inertia init my_awesome_project
```

### Options

```
      --git.remote string   git remote to use for continuous deployment (default "origin")
  -g, --global              just initialize global inertia configuration
  -h, --help                help for init
```

### Options inherited from parent commands

```
      --config string   specify relative path to Inertia project configuration (default "inertia.toml")
      --simple          disable colour and emoji output
```

### SEE ALSO

* [inertia](inertia.md)	 - Effortless, self-hosted continuous deployment for small teams and projects

