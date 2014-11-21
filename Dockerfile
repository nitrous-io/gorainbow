From ubuntu:14.04
MAINTAINER Nitrous.IO <eng@nitrous.io>

RUN apt-get update && apt-get install -y \
  build-essential \
  curl \
  git \
  mercurial \
  postgresql \
  ruby1.9.1 \
  sudo

RUN /etc/init.d/postgresql start \
  && sudo -u postgres createuser -U postgres --superuser root \
  && sudo -u postgres createdb -U postgres -O root -E utf8 testdb -T template0

RUN curl -s https://storage.googleapis.com/golang/go1.2.2.linux-amd64.tar.gz \
  | tar -v -C /usr/local -xz

RUN mkdir -p /opt/src/github.com/nitrous-io/gorainbow

ENV PATH /usr/local/go/bin:$PATH
ENV GOPATH /opt/src/github.com/nitrous-io/gorainbow/Godeps/_workspace:/opt

WORKDIR /opt/src/github.com/nitrous-io/gorainbow
ADD . /opt/src/github.com/nitrous-io/gorainbow
