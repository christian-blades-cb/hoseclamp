# Hoseclamp
[![](https://badge.imagelayers.io/christianbladescb/hoseclamp:latest.svg)](https://imagelayers.io/?images=christianbladescb/hoseclamp:latest 'Get your own badge on imagelayers.io')

Takes a stream from docker-loghose and squirts it into log.io.

## Usage

`docker run -v /var/run/docker.sock:/var/run/docker.sock --name=hoseclamp christianbladescb/hoseclamp -s LOGIOHOSTNAME:PORT`

## Why not use logstash?

It insulted my sister once.

## Other projects

* [docker-logio](https://github.com/gerchardon/docker-logio)
