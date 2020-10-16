// Copyright 2016 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// etcdctl is a command line application that controls etcd.
package main

import (
	"fmt"
	"os"

	"github.com/friendlyhank/etcd-3.4-annotated/etcdctl/ctlv2"
	"github.com/friendlyhank/etcd-3.4-annotated/etcdctl/ctlv3"
)

const (
	apiEnv = "ETCDCTL_API"
)

//客户端入口
func main() {
	apiv := os.Getenv(apiEnv)
	// unset apiEnv to avoid side-effect for future env and flag parsing.
	os.Unsetenv(apiEnv)
	//根据配置去读取V3还是V2
	if len(apiv) == 0 || apiv == "3" {
		ctlv3.Start()
		return
	}

	if apiv == "2" {
		ctlv2.Start()
		return
	}

	fmt.Fprintln(os.Stderr, "unsupported API version", apiv)
	os.Exit(1)
}
