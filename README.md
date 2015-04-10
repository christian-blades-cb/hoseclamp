# Hoseclamp

Takes a stream from docker-loghose and squirts it into log.io.

## Usage

`docker run -v /var/run/docker.sock:/var/run/docker.sock --name=hoseclamp christianbladescb/hoseclamp -s LOGIOHOSTNAME:PORT`

## Why not use logstash?

It insulted my sister once.

## Other projects

* [docker-logio](https://github.com/gerchardon/docker-logio)
