// Copyright 2015 The etcd Authors
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

package etcdhttp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"hank.com/etcd-3.3.12-annotated/etcdserver"
	"hank.com/etcd-3.3.12-annotated/etcdserver/api"
	"hank.com/etcd-3.3.12-annotated/etcdserver/api/membership"
	"hank.com/etcd-3.3.12-annotated/etcdserver/api/rafthttp"
	"hank.com/etcd-3.3.12-annotated/lease/leasehttp"
	"hank.com/etcd-3.3.12-annotated/pkg/types"

	"go.uber.org/zap"
)

const (
	peerMembersPath         = "/members"
	peerMemberPromotePrefix = "/members/promote/"
)

// NewPeerHandler generates an http.Handler to handle etcd peer requests.
//etcdserver handle
func NewPeerHandler(lg *zap.Logger, s etcdserver.ServerPeer) http.Handler {
	return newPeerHandler(lg, s, s.RaftHandler(), s.LeaseHandler())
}

//mewPeerHandler
func newPeerHandler(lg *zap.Logger, s etcdserver.Server, raftHandler http.Handler, leaseHandler http.Handler) http.Handler {
	//创建peerMember handers
	peerMembersHandler := newPeerMembersHandler(lg, s.Cluster())
	//创建peerMemberPromote handler
	peerMemberPromoteHandler := newPeerMemberPromoteHandler(lg, s)

	mux := http.NewServeMux()
	//路由 到对应的HandleFunc
	mux.HandleFunc("/", http.NotFound)
	mux.Handle(rafthttp.RaftPrefix, raftHandler)
	mux.Handle(rafthttp.RaftPrefix+"/", raftHandler)
	mux.Handle(peerMembersPath, peerMembersHandler)
	mux.Handle(peerMemberPromotePrefix, peerMemberPromoteHandler)
	if leaseHandler != nil {
		mux.Handle(leasehttp.LeasePrefix, leaseHandler)
		mux.Handle(leasehttp.LeaseInternalPrefix, leaseHandler)
	}
	//对应的方法
	mux.HandleFunc(versionPath, versionHandler(s.Cluster(), serveVersion))
	return mux
}

func newPeerMembersHandler(lg *zap.Logger, cluster api.Cluster) http.Handler {
	return &peerMembersHandler{
		lg:      lg,
		cluster: cluster,
	}
}

type peerMembersHandler struct {
	lg      *zap.Logger
	cluster api.Cluster
}

func newPeerMemberPromoteHandler(lg *zap.Logger, s etcdserver.Server) http.Handler {
	return &peerMemberPromoteHandler{
		lg:      lg,
		cluster: s.Cluster(),
		server:  s,
	}
}

type peerMemberPromoteHandler struct {
	lg      *zap.Logger
	cluster api.Cluster
	server  etcdserver.Server
}

/* 获取集群成员列表 */
func (h *peerMembersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, "GET") {
		return
	}
	w.Header().Set("X-Etcd-Cluster-ID", h.cluster.ID().String())

	if r.URL.Path != peerMembersPath {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	ms := h.cluster.Members() //获取集群成员列表
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ms); err != nil { //调用json接口,进行格式化并且将数据写到缓冲区中
		if h.lg != nil {
			h.lg.Warn("failed to encode membership members", zap.Error(err))
		} else {
			plog.Warningf("failed to encode members response (%v)", err)
		}
	}
}

func (h *peerMemberPromoteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !allowMethod(w, r, "POST") {
		return
	}
	w.Header().Set("X-Etcd-Cluster-ID", h.cluster.ID().String())

	if !strings.HasPrefix(r.URL.Path, peerMemberPromotePrefix) {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, peerMemberPromotePrefix)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("member %s not found in cluster", idStr), http.StatusNotFound)
		return
	}

	resp, err := h.server.PromoteMember(r.Context(), id)
	if err != nil {
		switch err {
		case membership.ErrIDNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case membership.ErrMemberNotLearner:
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
		case etcdserver.ErrLearnerNotReady:
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
		default:
			WriteError(h.lg, w, r, err)
		}
		if h.lg != nil {
			h.lg.Warn(
				"failed to promote a member",
				zap.String("member-id", types.ID(id).String()),
				zap.Error(err),
			)
		} else {
			plog.Errorf("error promoting member %s (%v)", types.ID(id).String(), err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		if h.lg != nil {
			h.lg.Warn("failed to encode members response", zap.Error(err))
		} else {
			plog.Warningf("failed to encode members response (%v)", err)
		}
	}
}
