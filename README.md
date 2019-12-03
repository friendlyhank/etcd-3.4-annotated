# etcd

[![Go Report Card](https://goreportcard.com/badge/github.com/etcd-io/etcd?style=flat-square)](https://goreportcard.com/report/github.com/etcd-io/etcd)
[![Coverage](https://codecov.io/gh/etcd-io/etcd/branch/master/graph/badge.svg)](https://codecov.io/gh/etcd-io/etcd)
[![Build Status Travis](https://img.shields.io/travis/etcd-io/etcdlabs.svg?style=flat-square&&branch=master)](https://travis-ci.com/etcd-io/etcd)
[![Build Status Semaphore](https://semaphoreci.com/api/v1/etcd-io/etcd/branches/master/shields_badge.svg)](https://semaphoreci.com/etcd-io/etcd)
[![Docs](https://img.shields.io/badge/docs-latest-green.svg)](https://etcd.io/docs)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/etcd-io/etcd)
[![Releases](https://img.shields.io/github/release/etcd-io/etcd/all.svg?style=flat-square)](https://github.com/etcd-io/etcd/releases)
[![LICENSE](https://img.shields.io/github/license/etcd-io/etcd.svg?style=flat-square)](https://github.com/etcd-io/etcd/blob/master/LICENSE)

**Note**: The `master` branch may be in an *unstable or even broken state* during development. Please use [releases][github-release] instead of the `master` branch in order to get stable binaries.

![etcd Logo](logos/etcd-horizontal-color.svg)

etcd is a distributed reliable key-value store for the most critical data of a distributed system, with a focus on being:

* *Simple*: well-defined, user-facing API (gRPC)
* *Secure*: automatic TLS with optional client cert authentication
* *Fast*: benchmarked 10,000 writes/sec
* *Reliable*: properly distributed using Raft

etcd is written in Go and uses the [Raft][raft] consensus algorithm to manage a highly-available replicated log.

etcd is used [in production by many companies](./ADOPTERS.md), and the development team stands behind it in critical deployment scenarios, where etcd is frequently teamed with applications such as [Kubernetes][k8s], [locksmith][locksmith], [vulcand][vulcand], [Doorman][doorman], and many others. Reliability is further ensured by [**rigorous testing**](https://github.com/etcd-io/etcd/tree/master/functional).

See [etcdctl][etcdctl] for a simple command line client.

[raft]: https://raft.github.io/
[k8s]: http://kubernetes.io/
[doorman]: https://github.com/youtube/doorman
[locksmith]: https://github.com/coreos/locksmith
[vulcand]: https://github.com/vulcand/vulcand
[etcdctl]: https://github.com/etcd-io/etcd/tree/master/etcdctl

## Community meetings

etcd contributors and maintainers have monthly (every four weeks) meetings at 11:00 AM (USA Pacific) on Thursday.

An initial agenda will be posted to the [shared Google docs][shared-meeting-notes] a day before each meeting, and everyone is welcome to suggest additional topics or other agendas.

[shared-meeting-notes]: https://docs.google.com/document/d/16XEGyPBisZvmmoIHSZzv__LoyOeluC5a4x353CX0SIM/edit


Time:
- [Jan 10th, 2019 11:00 AM video](https://www.youtube.com/watch?v=0Cphtbd1OSc&feature=youtu.be)
- [Feb 7th, 2019 11:00 AM video](https://youtu.be/U80b--oAlYM)
- [Mar 7th, 2019 11:00 AM video](https://youtu.be/w9TI5B7D1zg)
- [Apr 4th, 2019 11:00 AM video](https://youtu.be/oqQR2XH1L_A)
- [May 2nd, 2019 11:00 AM video](https://youtu.be/wFwQePuDWVw)
- [May 30th, 2019 11:00 AM video](https://youtu.be/2t1R5NATYG4)
- [Jul 11th, 2019 11:00 AM video](https://youtu.be/k_FZEipWD6Y)
- [Jul 25, 2019 11:00 AM video](https://youtu.be/VSUJTACO93I)
- [Aug 22, 2019 11:00 AM video](https://youtu.be/6IBQ-VxQmuM) 
- [Sep 19, 2019 11:00 AM video](https://youtu.be/SqfxU9DhBOc)
- Nov 14, 2019 11:00 AM
- Dec 12, 2019 11:00 AM

Join Hangouts Meet: [meet.google.com/umg-nrxn-qvs](https://meet.google.com/umg-nrxn-qvs)

Join by phone: +1 405-792-0633‬ PIN: ‪299 906‬#


## Getting started

### Getting etcd

The easiest way to get etcd is to use one of the pre-built release binaries which are available for OSX, Linux, Windows, and Docker on the [release page][github-release].

```json
{"errorCode":100,"message":"Key Not Found","cause":"/foo"}
```

### Watching a prefix

For those wanting to try the very latest version, [build the latest version of etcd][dl-build] from the `master` branch. This first needs [*Go*](https://golang.org/) installed (version 1.13+ is required). All development occurs on `master`, including new features and bug fixes. Bug fixes are first targeted at `master` and subsequently ported to release branches, as described in the [branch management][branch-management] guide.

[github-release]: https://github.com/etcd-io/etcd/releases
[branch-management]: ./Documentation/branch_management.md
[dl-build]: ./Documentation/dl_build.md#build-the-latest-version

### Running etcd

First start a single-member cluster of etcd.

If etcd is installed using the [pre-built release binaries][github-release], run it from the installation location as below:

```bash
/tmp/etcd-download-test/etcd
```

The etcd command can be simply run as such if it is moved to the system path as below:

```bash
mv /tmp/etcd-download-test/etcd /usr/local/bin/
etcd
```

If etcd is [built from the master branch][dl-build], run it as below:

```bash
./bin/etcd
```

This will bring up etcd listening on port 2379 for client communication and on port 2380 for server-to-server communication.

Next, let's set a single key, and then retrieve it:

Etcd can be used as a centralized coordination service in a cluster and `TestAndSet` is the most basic operation to build distributed lock service. This command will set the value only if the client provided `prevValue` is equal the current key value.

Here is a simple example. Let's create a key-value pair first: `testAndSet=one`.

```sh
curl -L http://127.0.0.1:4001/v1/keys/testAndSet -d value=one
```
etcdctl put mykey "this is awesome"
etcdctl get mykey
```

etcd is now running and serving client requests. For more, please check out:

- [Interactive etcd playground](http://play.etcd.io)
- [Animated quick demo](./Documentation/demo.md)

### etcd TCP ports

The [official etcd ports][iana-ports] are 2379 for client requests, and 2380 for peer communication.

[iana-ports]: http://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.txt

### Running a local etcd cluster

First install [goreman](https://github.com/mattn/goreman), which manages Procfile-based applications.

Our [Procfile script](./Procfile) will set up a local example cluster. Start it with:

```bash
goreman start
```

Now list the keys under `/foo`

Every cluster member and proxy accepts key value reads and key value writes.

Follow the steps in [Procfile.learner](./Procfile.learner) to add a learner node to the cluster. Start the learner node with:

```bash
goreman -f ./Procfile.learner start
```

### Next steps

Now it's time to dig into the full etcd API and other guides.

- Read the full [documentation][fulldoc].
- Explore the full gRPC [API][api].
- Set up a [multi-machine cluster][clustering].
- Learn the [config format, env variables and flags][configuration].
- Find [language bindings and tools][integrations].
- Use TLS to [secure an etcd cluster][security].
- [Tune etcd][tuning].

[fulldoc]: ./Documentation/docs.md
[api]: ./Documentation/dev-guide/api_reference_v3.md
[clustering]: ./Documentation/op-guide/clustering.md
[configuration]: ./Documentation/op-guide/configuration.md
[integrations]: ./Documentation/integrations.md
[security]: ./Documentation/op-guide/security.md
[tuning]: ./Documentation/tuning.md

First, you need to have a CA cert `clientCA.crt` and signed key pair `client.crt`, `client.key`. This site has a good reference for how to generate self-signed key pairs:

```url
http://www.g-loaded.eu/2005/11/10/be-your-own-ca/
```

Next, lets configure etcd to use this keypair:

```sh
./etcd -clientCert client.crt -clientKey client.key -f
```

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches and the contribution workflow.

You can now test the configuration using https:

## Reporting a security vulnerability

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

### Authentication with HTTPS client certificates

## Issue and PR management

See [issue triage guidelines](Documentation/triage/issues.md) for details on how issues are managed.

Try the same request to this server:

## etcd Emeritus Maintainers

The request should be rejected by the server.

```
...
routines:SSL3_READ_BYTES:sslv3 alert bad certificate
...
```

We need to give the CA signed cert to the server.

```sh
curl -L https://127.0.0.1:4001/v1/keys/foo -d value=bar -v --key myclient.key --cert myclient.crt -cacert clientCA.crt
```

You should able to see
```
...
SSLv3, TLS handshake, CERT verify (15):
...
TLS handshake, Finished (20)
```

And also the response from the server:

```json
{"action":"SET","key":"/foo","value":"bar","newKey":true,"index":3}
```

## Clustering

### Example cluster of three machines

Let's explore the use of etcd clustering. We use go-raft as the underlying distributed protocol which provides consistency and persistence of the data across all of the etcd instances.

Let start by creating 3 new etcd instances.

We use -s to specify server port and -c to specify client port and -d to specify the directory to store the log and info of the node in the cluster

```sh
./etcd -s 7001 -c 4001 -d nodes/node1
```

Let the join two more nodes to this cluster using the -C argument:

```sh
./etcd -c 4002 -s 7002 -C 127.0.0.1:7001 -d nodes/node2
./etcd -c 4003 -s 7003 -C 127.0.0.1:7001 -d nodes/node3
```

Get the machines in the cluster:

```sh
curl -L http://127.0.0.1:4001/machines
```

We should see there are three nodes in the cluster

```
0.0.0.0:4001,0.0.0.0:4002,0.0.0.0:4003
```

The machine list is also available via this API:

```sh
curl -L http://127.0.0.1:4001/v1/keys/_etcd/machines
```

```json
[{"action":"GET","key":"/machines/node1","value":"0.0.0.0,7001,4001","index":4},{"action":"GET","key":"/machines/node3","value":"0.0.0.0,7002,4002","index":4},{"action":"GET","key":"/machines/node4","value":"0.0.0.0,7003,4003","index":4}]
```

The key of the machine is based on the ```commit index``` when it was added. The value of the machine is ```hostname```, ```raft port``` and ```client port```.

Also try to get the current leader in the cluster

```
curl -L http://127.0.0.1:4001/leader
```
The first server we set up should be the leader, if it has not dead during these commands.

```
0.0.0.0:7001
```

Now we can do normal SET and GET operations on keys as we explored earlier.

```sh
curl -L http://127.0.0.1:4001/v1/keys/foo -d value=bar
```

```json
{"action":"SET","key":"/foo","value":"bar","newKey":true,"index":5}
```

### Killing Nodes in the Cluster

Let's kill the leader of the cluster and get the value from the other machine:

```sh
curl -L http://127.0.0.1:4002/v1/keys/foo
```

A new leader should have been elected.

```
curl -L http://127.0.0.1:4001/leader
```

```
0.0.0.0:7002 or 0.0.0.0:7003
```

You should be able to see this:

```json
{"action":"GET","key":"/foo","value":"bar","index":5}
```

It succeeded!

### Testing Persistence

OK. Next let us kill all the nodes to test persistence. And restart all the nodes use the same command as before.

Your request for the `foo` key will return the correct value:

```sh
curl -L http://127.0.0.1:4002/v1/keys/foo
```

```json
{"action":"GET","key":"/foo","value":"bar","index":5}
```

* Fanmin Shi 
* Anthony Romano 

In the previous example we showed how to use SSL client certs for client to server communication. Etcd can also do internal server to server communication using SSL client certs. To do this just change the ```-client*``` flags to ```-server*```.

If you are using SSL for server to server communication, you must use it on all instances of etcd.
