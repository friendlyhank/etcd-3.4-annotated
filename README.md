# etcd

## Getting Started

### Setting up a node

```sh
./etcd
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

```sh
curl http://127.0.0.1:4001/v1/keys/message
```

You should receive the response as

```json
{"action":"GET","key":"/message","value":"Hello world","index":3}
```
#### Changing the value of a key

etcd is used [in production by many companies](./ADOPTERS.md), and the development team stands behind it in critical deployment scenarios, where etcd is frequently teamed with applications such as [Kubernetes][k8s], [locksmith][locksmith], [vulcand][vulcand], [Doorman][doorman], and many others. Reliability is further ensured by [**rigorous testing**](https://github.com/etcd-io/etcd/tree/master/functional).

See [etcdctl][etcdctl] for a simple command line client.

[raft]: https://raft.github.io/
[k8s]: http://kubernetes.io/
[doorman]: https://github.com/youtube/doorman
[locksmith]: https://github.com/coreos/locksmith
[vulcand]: https://github.com/vulcand/vulcand
[etcdctl]: https://github.com/etcd-io/etcd/tree/master/etcdctl

There is a new field in the response: prevValue. It is the value of the key before the change happened.

etcd contributors and maintainers have monthly (every four weeks) meetings at 11:00 AM (USA Pacific) on Thursday.

An initial agenda will be posted to the [shared Google docs][shared-meeting-notes] a day before each meeting, and everyone is welcome to suggest additional topics or other agendas.

You should see the response as

```json
{"action":"DELETE","key":"/message","prevValue":"Hello etcd","index":5}
```

#### Using time to live key

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

```json
{"action":"SET","key":"/foo","value":"bar","newKey":true,"expiration":"2013-07-11T20:31:12.156146039-07:00","ttl":4,"index":6}
```

There are the last two new fields in response.

Expiration field is the time that this key will expire and be deleted.


## Getting started

```sh
curl http://127.0.0.1:4001/v1/keys/foo
```
You can expect the ttl is counting down and after 5 seconds you should see this,

The easiest way to get etcd is to use one of the pre-built release binaries which are available for OSX, Linux, Windows, and Docker on the [release page][github-release].

which indicates the key has expired and was deleted.

For those wanting to try the very latest version, [build the latest version of etcd][dl-build] from the `master` branch. This first needs [*Go*](https://golang.org/) installed (version 1.13+ is required). All development occurs on `master`, including new features and bug fixes. Bug fixes are first targeted at `master` and subsequently ported to release branches, as described in the [branch management][branch-management] guide.

[github-release]: https://github.com/etcd-io/etcd/releases
[branch-management]: ./Documentation/branch_management.md
[dl-build]: ./Documentation/dl_build.md#build-the-latest-version

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

If etcd is [built from the master branch][dl-build], run it as below:

```bash
./bin/etcd
```

This will bring up etcd listening on port 2379 for client communication and on port 2380 for server-to-server communication.

Next, let's set a single key, and then retrieve it:

What it does is to test whether the given previous value is equal to the value of the key, if equal etcd will change the value of the key to the given value.

Here is a simple example.
Let us create a key-value pair first: `testAndSet=one`.

```sh
curl http://127.0.0.1:4001/v1/keys/testAndSet -d value=one
```

Let us try a invaild `TestAndSet` command.

```sh
curl http://127.0.0.1:4001/v1/testAndSet/testAndSet -d prevValue=two -d value=three
```

etcd is now running and serving client requests. For more, please check out:

The response should be

```html
Test one==two fails
```

which means `testAndSet` failed.

Let us try a vaild one.

```sh
curl http://127.0.0.1:4001/v1/testAndSet/testAndSet -d prevValue=one -d value=two
```

The response should be

```json
{"action":"SET","key":"/testAndSet","prevValue":"one","value":"two","index":10}
```

The [official etcd ports][iana-ports] are 2379 for client requests, and 2380 for peer communication.

[iana-ports]: http://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.txt

#### Listing directory

Last we provide a simple List command to list all the keys under a prefix path.

First install [goreman](https://github.com/mattn/goreman), which manages Procfile-based applications.

Our [Procfile script](./Procfile) will set up a local example cluster. Start it with:

```bash
goreman start
```

This will bring up 3 etcd members `infra1`, `infra2` and `infra3` and etcd `grpc-proxy`, which runs locally and composes a cluster.

Every cluster member and proxy accepts key value reads and key value writes.

We should see the response as

```bash
goreman -f ./Procfile.learner start
```

which meas `foo=barbar` is a key-value pair under `/foo` and `foo_dir` is a directory.

### Setting up a cluster of three machines

Next we can explore the power of etcd cluster. We use go-raft as the underlay distributed protocol which provide consistency and persistence of all the machines in the cluster. The will allow if the minor machine dies, the cluster will still be able to performance correctly. Also if most of the machines dead and restart,  we will recover from the previous state of the cluster.

Now it's time to dig into the full etcd API and other guides.

The first one will be

[fulldoc]: ./Documentation/docs.md
[api]: ./Documentation/dev-guide/api_reference_v3.md
[clustering]: ./Documentation/op-guide/clustering.md
[configuration]: ./Documentation/op-guide/configuration.md
[integrations]: ./Documentation/integrations.md
[security]: ./Documentation/op-guide/security.md
[tuning]: ./Documentation/tuning.md

## Contact

- Mailing list: [etcd-dev](https://groups.google.com/forum/?hl=en#!forum/etcd-dev)
- IRC: #[etcd](irc://irc.freenode.org:6667/#etcd) on freenode.org
- Planning/Roadmap: [milestones](https://github.com/etcd-io/etcd/milestones), [roadmap](./ROADMAP.md)
- Bugs: [issues](https://github.com/etcd-io/etcd/issues)

Let the second one join it.

```sh
./etcd -c 4002 -s 7002 -C 127.0.0.1:7001 -d nod/node2
```

And the third one:

```sh
./etcd -c 4003 -s 7003 -C 127.0.0.1:7001 -d nod/node3
```

## Reporting bugs

See [reporting bugs](Documentation/reporting_bugs.md) for details about reporting any issues.

## Reporting a security vulnerability

See [security disclosure and release process](security/README.md) for details on how to report a security vulnerability and how the etcd team manages it.

## Issue and PR management

See [issue triage guidelines](Documentation/triage/issues.md) for details on how issues are managed.

See [PR management](Documentation/triage/PRs.md) for guidelines on how pull requests are managed.

## etcd Emeritus Maintainers

These emeritus maintainers dedicated a part of their career to etcd and reviewed code, triaged bugs, and pushed the project forward over a substantial period of time. Their contribution is greatly appreciated.

* Fanmin Shi 
* Anthony Romano 

### License

Try

```sh
curl http://127.0.0.1:4002/v1/keys/foo
```
You should able to see

```json
{"action":"GET","key":"/foo","value":"bar","index":5}
```
