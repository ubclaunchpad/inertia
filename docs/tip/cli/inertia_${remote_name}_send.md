## inertia ${remote_name} send

Send a file to your Inertia deployment

### Synopsis

Sends a file, such as a configuration or .env file, to your Inertia deployment.

```
inertia ${remote_name} send [filepath] [flags]
```

### Options

```
  -d, --dest string   path relative from project root to send file to
  -h, --help          help for send
  -p, --perm string   permissions settings to create file with (default "0655")
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

