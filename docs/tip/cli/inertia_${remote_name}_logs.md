## inertia ${remote_name} logs

Access logs of containers on your remote host

### Synopsis

Accesses logs of containers on your remote host.
	
By default, this command retrieves Inertia daemon logs, but you can provide an
argument that specifies the name of the container you wish to retrieve logs for.
Use 'inertia [remote] status' to see which containers are active.

```
inertia ${remote_name} logs [container] [flags]
```

### Options

```
      --entries int   Number of log entries to fetch
  -h, --help          help for logs
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

