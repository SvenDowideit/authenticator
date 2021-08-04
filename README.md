# Authenticator

A tool to help you populate your `config.json` with the Portainer API Authorization header, and populate Docker contexts.

## Usage

```
$ docker run --rm -v ~/.docker/:/.docker/ portainer/authenticator --context --config /.docker/config.json http://PORTAINER_URL:9000 username password
```

or if running locally:

```
./main --context --config ~/.docker/config.json http://PORTAINER_URL:9000 username password
```

This will add the jwt header to the docker cli header config, and install all DOcker based endpoints as Docker contexts
## Docker CLI + Portainer API

using a Docker context (see what they are using `docker context ls`):

```
$ docker --context <entrypointname> ps -a                                       

CONTAINER ID        IMAGE                 COMMAND                  CREATED             STATUS                         PORTS                                            NAMES
04e273b9cb27        portainer/base        "/app/portainer --no…"   5 minutes ago       Up 5 minutes                   0.0.0.0:9000->9000/tcp   portainer
...
```

```
$ docker -H PORTAINER_URL:9000/api/endpoints/1/docker ps -a                                       

CONTAINER ID        IMAGE                 COMMAND                  CREATED             STATUS                         PORTS                                            NAMES
04e273b9cb27        portainer/base        "/app/portainer --no…"   5 minutes ago       Up 5 minutes                   0.0.0.0:9000->9000/tcp   portainer
...
```
