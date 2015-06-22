FROM centurylink/ca-certs

MAINTAINER Christian Blades <christian.blades@gmail.com>

COPY hoseclamp /
ENTRYPOINT ["/hoseclamp"]
