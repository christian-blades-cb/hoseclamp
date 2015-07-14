# Hoseclamp
[![](https://badge.imagelayers.io/christianbladescb/hoseclamp:latest.svg)](https://imagelayers.io/?images=christianbladescb/hoseclamp:latest 'Get your own badge on imagelayers.io')

Takes a stream from docker-loghose and squirts it into log.io.

## Usage

```shell
Usage:
  hoseclamp [OPTIONS]

Application Options:
  -s, --server=   logio server (localhost:28777) [$LOGIO_SERVER]
  -u, --sumourl=  http collector endpoint for Sumologic [$SUMOLOGIC_ENDPOINT]
      --host=     Docker Host (unix:///var/run/docker.sock) [$DOCKER_HOST]
      --certpath= Docker TLS Certificate path [$DOCKER_CERT_PATH]
  -v, --verbose   all the logs

Help Options:
  -h, --help      Show this help message
```

### Example

`docker run -v /var/run/docker.sock:/var/run/docker.sock --name=hoseclamp -e SUMOLOGIC_ENDPOINT=https://sumologic.api/endpoint christianbladescb/hoseclamp`

## Why not use logstash?

It insulted my sister once.

## Other projects

* [docker-logio](https://github.com/gerchardon/docker-logio)
