# Grawp #
## A Minecraft Server management system. ##
Allows administrators to preconfigure and build container
services for the purpose of running a server ecosystem.

## Capabilities ##
The application has the following capabilities for admin
convenience:

- Archiving service assets
- Creating service images
- Creating service containers
- Running service containers with "watchdog" restarts

```bash
$ ./grawpa
Grawpadmin is an application meant for maintaining this project and its processes

Usage:
  grawpadmin [command]

Available Commands:
  archive      Create tar ball(s) of server assets.
  completion   Generate the autocompletion script for the specified shell
  help         Help about any command
  images       manage service images
  manifest     Print the service manifest to stdout
  services     manage service containers
  watch        Start and watch a running service container

Flags:
  -h, --help      help for grawpadmin
  -v, --version   version for grawpadmin

Use "grawpadmin [command] --help" for more information about a command.
```
