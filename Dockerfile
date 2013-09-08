from ubuntu:12.04
maintainer Scott Nelson "scott@scttnlsn.com"

run apt-get update
run apt-get install -y python-software-properties git wget build-essential

# Go
run add-apt-repository -y ppa:duh/golang
run apt-get update
run apt-get install -y golang
run mkdir /go
env GOPATH /go

# LevelDB
run wget https://leveldb.googlecode.com/files/leveldb-1.13.0.tar.gz --no-check-certificate
run tar -zxvf leveldb-1.13.0.tar.gz
run cd leveldb-1.13.0; make
run cp -r leveldb-1.13.0/include/leveldb /usr/include/
run cp leveldb-1.13.0/libleveldb.* /usr/lib/

# Queued
run mkdir -p /queued/src
run git clone https://github.com/scttnlsn/queued.git /queued/src
run go get github.com/jmhodges/levigo
run go get github.com/gorilla/mux
run cd /queued/src; make
run cp /queued/src/build/queued /usr/bin/queued

expose 5353
entrypoint ["/usr/bin/queued", "-db-path=/queued/db"]