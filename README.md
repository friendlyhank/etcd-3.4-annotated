# etcd

A highly-available key value store for shared configuration and service discovery. etcd is inspired by zookeeper and doozer, with a focus on:

* Simple: curl'able user facing API (HTTP+JSON)
* Secure: optional SSL client cert authentication
* Fast: benchmarked 1000s of writes/s per instance
* Reliable: Properly distributed using Raft

Etcd is written in go and uses the [raft][raft] consensus algorithm to a manage replicated
log for high availability. 

See [go-etcd][go-etcd] for a native go client. Or feel free to just use curl, as in the examples below. 

[raft]: https://github.com/coreos/go-raft
[go-etcd]: https://github.com/coreos/go-etcd

## Getting Started

### Building

etcd is installed like any other Go binary. The steps below will put everything into a directory called etcd.

```
mkdir etcd
cd etcd
export GOPATH=`pwd`
go get github.com/coreos/etcd
go install github.com/coreos/etcd
```

### Running a single node

These examples will use a single node cluster to show you the basics of the etcd REST API. Lets start etcd:

```sh
./bin/etcd
```

This will bring up a node, which will be listening on internal port 7001 (for server communication) and external port 4001 (for client communication)

#### Setting the value to a key

Let’s set the first key-value pair to the node. In this case the key is `/message` and the value is `Hello world`.

```sh
curl http://127.0.0.1:4001/v1/keys/message -d value="Hello world"
```

```json
{"action":"SET","key":"/message","value":"Hello world","newKey":true,"index":3}
```

This response contains five fields. We will introduce three more fields as we try more commands.

1. The action of the request; we set the value via a POST request, thus the action is `SET`.
 
2. The key of the request; we set `/message` to `Hello world!`, so the key field is `/message`.
Notice we use a file system like structure to represent the key-value pairs. So each key starts with `/`.

[![Go Report Card](https://goreportcard.com/badge/github.com/etcd-io/etcd?style=flat-square)](https://goreportcard.com/report/github.com/etcd-io/etcd)
[![Coverage](https://codecov.io/gh/etcd-io/etcd/branch/master/graph/badge.svg)](https://codecov.io/gh/etcd-io/etcd)
[![Build Status Travis](https://img.shields.io/travis/etcd-io/etcdlabs.svg?style=flat-square&&branch=master)](https://travis-ci.com/etcd-io/etcd)
[![Build Status Semaphore](https://semaphoreci.com/api/v1/etcd-io/etcd/branches/master/shields_badge.svg)](https://semaphoreci.com/etcd-io/etcd)
[![Docs](https://img.shields.io/badge/docs-latest-green.svg)](https://etcd.io/docs)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/etcd-io/etcd)
[![Releases](https://img.shields.io/github/release/etcd-io/etcd/all.svg?style=flat-square)](https://github.com/etcd-io/etcd/releases)
[![LICENSE](https://img.shields.io/github/license/etcd-io/etcd.svg?style=flat-square)](https://github.com/etcd-io/etcd/blob/master/LICENSE)

**Note**: The `master` branch may be in an *unstable or even broken state* during development. Please use [releases][github-release] instead of the `master` branch in order to get stable binaries.

5. Index field is the unique request index of the set request. Each sensitive request we send to the server will have a unique request index. The current sensitive request are `SET`, `DELETE` and `TESTANDSET`. All of these request will change the state of the key-value store system, thus they are sensitive. `GET`, `LIST` and `WATCH` are non-sensitive commands. Those commands will not change the state of the key-value store system. You may notice that in this example the index is 3, although it is the first request you sent to the server. This is because there are some internal commands that also change the state of the server, we also need to assign them command indexes(Command used to add a server and sync the servers).

etcd is a distributed reliable key-value store for the most critical data of a distributed system, with a focus on being:

Get the value that we just set in `/message` by issuing a GET:

```sh
curl http://127.0.0.1:4001/v1/keys/message
```

```json
{"action":"GET","key":"/message","value":"Hello world","index":3}
```
#### Changing the value of a key

Change the value of `/message` from `Hello world` to `Hello etcd` with another POST to the key:

See [etcdctl][etcdctl] for a simple command line client.

[raft]: https://raft.github.io/
[k8s]: http://kubernetes.io/
[doorman]: https://github.com/youtube/doorman
[locksmith]: https://github.com/coreos/locksmith
[vulcand]: https://github.com/vulcand/vulcand
[etcdctl]: https://github.com/etcd-io/etcd/tree/master/etcdctl

Notice that the `prevValue` is set to `Hello world`.

etcd contributors and maintainers have monthly (every four weeks) meetings at 11:00 AM (USA Pacific) on Thursday.

Remove the `/message` key with a DELETE:

```sh
curl http://127.0.0.1:4001/v1/keys/message -X DELETE
```

```json
{"action":"DELETE","key":"/message","prevValue":"Hello etcd","index":5}
```

#### Using a TTL on a key

Keys in etcd can be set to expire after a specified number of seconds. That is done by setting a TTL (time to live) on the key when you POST:

```sh
curl http://127.0.0.1:4001/v1/keys/foo -d value=bar -d ttl=5
```

```json
{"action":"SET","key":"/foo","value":"bar","newKey":true,"expiration":"2013-07-11T20:31:12.156146039-07:00","ttl":4,"index":6}
```

Note the last two new fields in response:

1. The expiration is the time that this key will expire and be deleted.

2. The ttl is the time to live of the key.

Now you can try to get the key by sending:

```sh
curl http://127.0.0.1:4001/v1/keys/foo
```

If the TTL has expired, the key will be deleted, and you will be returned a 404.

The easiest way to get etcd is to use one of the pre-built release binaries which are available for OSX, Linux, Windows, and Docker on the [release page][github-release].


For those wanting to try the very latest version, [build the latest version of etcd][dl-build] from the `master` branch. This first needs [*Go*](https://golang.org/) installed (version 1.13+ is required). All development occurs on `master`, including new features and bug fixes. Bug fixes are first targeted at `master` and subsequently ported to release branches, as described in the [branch management][branch-management] guide.

We can watch a path prefix and get notifications if any key change under that prefix.

In one terminal, we send a watch request:

```sh
curl http://127.0.0.1:4001/v1/watch/foo
```

First start a single-member cluster of etcd.

If etcd is installed using the [pre-built release binaries][github-release], run it from the installation location as below:

```bash
/tmp/etcd-download-test/etcd
```

The first terminal should get the notification and return with the same response as the set request.

```bash
mv /tmp/etcd-download-test/etcd /usr/local/bin/
etcd
```

However, the watch command can do more than this. Using the the index we can watch for commands that has happened in the past. This is useful for ensuring you don't miss events between watch commands.

Let's try to watch for the set command of index 6 again:

```sh
curl http://127.0.0.1:4001/v1/watch/foo -d index=7
```

The watch command returns immediately with the same response as previous.

#### Atomic Test and Set

Etcd servers will process all the command in sequence atomically. Thus it can be used as a centralized coordination service in a cluster.

`TestAndSet` is the most basic operation to build distributed lock service.

The basic logic is to test whether the given previous value is equal to the value of the key, if equal etcd will change the value of the key to the given value.

Here is a simple example. Let's create a key-value pair first: `testAndSet=one`.

```sh
curl http://127.0.0.1:4001/v1/keys/testAndSet -d value=one
```

Let's try an invaild `TestAndSet` command.
We can give another parameter prevValue to set command to make it a TestAndSet command.

```sh
curl http://127.0.0.1:4001/v1/keys/testAndSet -d prevValue=two -d value=three
```

etcd is now running and serving client requests. For more, please check out:

```json
{"errorCode":101,"message":"The given PrevValue is not equal to the value of the key","cause":"TestAndSet: one!=two"}
```

which means `testAndSet` failed.

Let us try a vaild one.

```sh
curl http://127.0.0.1:4001/v1/keys/testAndSet -d prevValue=one -d value=two
```

The response should be

```json
{"action":"SET","key":"/testAndSet","prevValue":"one","value":"two","index":10}
```

We successfully changed the value from “one” to “two”, since we give the correct previous value.

[iana-ports]: http://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.txt

#### Listing directory

Last we provide a simple List command to list all the keys under a prefix path.

First install [goreman](https://github.com/mattn/goreman), which manages Procfile-based applications.

Our [Procfile script](./Procfile) will set up a local example cluster. Start it with:

```sh
curl http://127.0.0.1:4001/v1/keys/foo/foo_dir/bar -d value=barbarbar
```

This will bring up 3 etcd members `infra1`, `infra2` and `infra3` and etcd `grpc-proxy`, which runs locally and composes a cluster.

```sh
curl http://127.0.0.1:4001/v1/get/foo/
```

We should see the response as an array of items

```json
[{"action":"GET","key":"/foo/foo","value":"barbar","index":10},{"action":"GET","key":"/foo/foo_dir","dir":true,"index":10}]
```

which meas `foo=barbar` is a key-value pair under `/foo` and `foo_dir` is a directory.

#### Using HTTPS between server and client
Etcd supports SSL/TLS and client cert authentication for clients to server, as well as server to server communication

Before that we need to have a CA cert```clientCA.crt``` and signed key pair ```client.crt, client.key``` .

This site has a good reference for how to generate self-signed key pairs
```url
http://www.g-loaded.eu/2005/11/10/be-your-own-ca/
```

```sh
./etcd -clientCert client.crt -clientKey client.key -i
```

```-i``` is to ignore the previously created default configuration file.
```-clientCert``` and ```-clientKey``` are the key and cert for transport layer security between client and server

```sh
curl https://127.0.0.1:4001/v1/keys/foo -d value=bar -v -k
```

or 

```sh
curl https://127.0.0.1:4001/v1/keys/foo -d value=bar -v -cacert clientCA.crt
```

You should be able to see the handshake succeed.
```
...
SSLv3, TLS handshake, Finished (20):
...
```
And also the response from the etcd server.
```json
{"action":"SET","key":"/foo","value":"bar","newKey":true,"index":3}
```

We also can do authentication using CA cert. The clients will also need to provide their cert to the server. The server will check whether the cert is signed by the CA and decide whether to serve the request.

```sh
./etcd -clientCert client.crt -clientKey client.key -clientCAFile clientCA.crt -i
```

```-clientCAFile``` is the path to the CA cert.

Try the same request to this server.
```sh
curl https://127.0.0.1:4001/v1/keys/foo -d value=bar -v -k
```
or 

```sh
curl https://127.0.0.1:4001/v1/keys/foo -d value=bar -v -cacert clientCA.crt
```

The request should be rejected by the server.
```
...
routines:SSL3_READ_BYTES:sslv3 alert bad certificate
...
```

We need to give the CA signed cert to the server. 
```sh
curl https://127.0.0.1:4001/v1/keys/foo -d value=bar -v --key myclient.key --cert myclient.crt -k
```

or

```sh
curl https://127.0.0.1:4001/v1/keys/foo -d value=bar -v --key myclient.key --cert myclient.crt -cacert clientCA.crt
```

You should able to see
```
...
SSLv3, TLS handshake, CERT verify (15):
...
TLS handshake, Finished (20)
```

And also the response from the server
```json
{"action":"SET","key":"/foo","value":"bar","newKey":true,"index":3}
```

### Setting up a cluster of three machines

Next let's explore the use of etcd clustering. We use go-raft as the underlying distributed protocol which provides consistency and persistence of the data across all of the etcd instances.

Let start by creating 3 new etcd instances.

[fulldoc]: ./Documentation/docs.md
[api]: ./Documentation/dev-guide/api_reference_v3.md
[clustering]: ./Documentation/op-guide/clustering.md
[configuration]: ./Documentation/op-guide/configuration.md
[integrations]: ./Documentation/integrations.md
[security]: ./Documentation/op-guide/security.md
[tuning]: ./Documentation/tuning.md

## Contact

Let the join two more nodes to this cluster using the -C argument:

```sh
./etcd -c 4002 -s 7002 -C 127.0.0.1:7001 -d nod/node2
./etcd -c 4003 -s 7003 -C 127.0.0.1:7001 -d nod/node3
```

Get the machines in the cluster

```sh
curl http://127.0.0.1:4001/machines
```

We should see there are three nodes in the cluster

```
0.0.0.0:4001,0.0.0.0:4002,0.0.0.0:4003
```

Also try to get the current leader in the cluster

```
curl http://127.0.0.1:4001/leader
```
The first server we set up should be the leader, if it has not dead during these commands.

```
0.0.0.0:7001
```

Now we can do normal SET and GET operations on keys as we explored earlier.

See [reporting bugs](Documentation/reporting_bugs.md) for details about reporting any issues.

## Reporting a security vulnerability

#### Killing Nodes in the Cluster

Let's kill the leader of the cluster and get the value from the other machine:

See [PR management](Documentation/triage/PRs.md) for guidelines on how pull requests are managed.

A new leader should have been elected.

```
curl http://127.0.0.1:4001/leader
```

```
0.0.0.0:7002 or 0.0.0.0:7003
```

You should be able to see this:

These emeritus maintainers dedicated a part of their career to etcd and reviewed code, triaged bugs, and pushed the project forward over a substantial period of time. Their contribution is greatly appreciated.

* Fanmin Shi 
* Anthony Romano 

#### Testing Persistence

OK. Next let us kill all the nodes to test persistence. And restart all the nodes use the same command as before.

Your request for the `foo` key will return the correct value:

```sh
curl http://127.0.0.1:4002/v1/keys/foo
```

```json
{"action":"GET","key":"/foo","value":"bar","index":5}
```

#### Using HTTPS between servers
In the previous example we showed how to use SSL client certs for client to server communication. Etcd can also do internal server to server communication using SSL client certs. To do this just change the ```-client*``` flags to ```-server*```.
If you are using SSL for server to server communication, you must use it on all instances of etcd.

