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

package rafthttp

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/friendlyhank/etcd-3.4-annotated/etcdserver/api/snap"
	stats "github.com/friendlyhank/etcd-3.4-annotated/etcdserver/api/v2stats"
	"github.com/friendlyhank/etcd-3.4-annotated/pkg/logutil"
	"github.com/friendlyhank/etcd-3.4-annotated/pkg/transport"
	"github.com/friendlyhank/etcd-3.4-annotated/pkg/types"
	"github.com/friendlyhank/etcd-3.4-annotated/raft"
	"github.com/friendlyhank/etcd-3.4-annotated/raft/raftpb"

	"github.com/coreos/pkg/capnslog"
	"github.com/xiang90/probing"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

var plog = logutil.NewMergeLogger(capnslog.NewPackageLogger("github.com/friendlyhank/etcd-3.4-annotated", "rafthttp"))

type Raft interface {
	//将指定的消息实例传递到底层的etcd-raft模块进行处理
	Process(ctx context.Context, m raftpb.Message) error
	IsIDRemoved(id uint64) bool//检测指定节点是否从当前集群中移除
	ReportUnreachable(id uint64)//通知底层etcd-raft模块,当前节点与指定的节点无法连通
	ReportSnapshot(id uint64, status raft.SnapshotStatus)//通知底层etcd-raft模块,快照数据是否发送成功
}

type Transporter interface {
	// Start starts the given Transporter.
	// Start MUST be called before calling other functions in the interface.
	Start() error//初始化操作
	// Handler returns the HTTP handler of the transporter.
	// A transporter HTTP handler handles the HTTP requests
	// from remote peers.
	// The handler MUST be used to handle RaftPrefix(/raft)
	// endpoint.
	Handler() http.Handler//创建Handler实例，并关联到指定的URL上
	// Send sends out the given messages to the remote peers.
	// Each message has a To field, which is an id that maps
	// to an existing peer in the transport.
	// If the id cannot be found in the transport, the message
	// will be ignored.
	Send(m []raftpb.Message) //发送消息
	// SendSnapshot sends out the given snapshot message to a remote peer.
	// The behavior of SendSnapshot is similar to Send.
	SendSnapshot(m snap.Message)//发送快照数据
	// AddRemote adds a remote with given peer urls into the transport.
	// A remote helps newly joined member to catch up the progress of cluster,
	// and will not be used after that.
	// It is the caller's responsibility to ensure the urls are all valid,
	// or it panics.
	AddRemote(id types.ID, urls []string)//在集群中添加一个节点时，其他节点会通过该方法添加该新节点消息
	// AddPeer adds a peer with given peer urls into the transport.
	// It is the caller's responsibility to ensure the urls are all valid,
	// or it panics.
	// Peer urls are used to connect to the remote peer.
	AddPeer(id types.ID, urls []string)//添加指定集群节点(自己除外)
	// RemovePeer removes the peer with given id.
	RemovePeer(id types.ID) //移除集群指定节点
	// RemoveAllPeers removes all the existing peers in the transport.
	RemoveAllPeers()//移除所有的节点
	// UpdatePeer updates the peer urls of the peer with the given id.
	// It is the caller's responsibility to ensure the urls are all valid,
	// or it panics.
	UpdatePeer(id types.ID, urls []string)//修改节点信息
	// ActiveSince returns the time that the connection with the peer
	// of the given id becomes active.
	// If the connection is active since peer was added, it returns the adding time.
	// If the connection is currently inactive, it returns zero time.
	ActiveSince(id types.ID) time.Time//节点存活时长的统计
	// ActivePeers returns the number of active peers.
	ActivePeers() int//处于活跃状态得节点数
	// Stop closes the connections and stops the transporter.
	Stop()//关闭所有连接
}

// Transport implements Transporter interface. It provides the functionality
// to send raft messages to peers, and receive raft messages from peers.
// User should call Handler method to get a handler to serve requests
// received from peerURLs.
// User needs to call Start before calling other functions, and call
// Stop when the Transport is no longer used.
type Transport struct {
	Logger *zap.Logger

	DialTimeout time.Duration //拨号超时的时间 maximum duration before timing out dial of the request
	// DialRetryFrequency defines the frequency of streamReader dial retrial attempts;
	// a distinct rate limiter is created per every peer (default value: 10 events/sec)
	DialRetryFrequency rate.Limit//拨号重试的频率

	TLSInfo transport.TLSInfo //tls信息安全 TLS information used when creating connection

	ID          types.ID   //当前节点自己的ID local member ID
	URLs        types.URLs //当前节点与集群其他节点交互时使用的URL地址 local peer URLs
	ClusterID   types.ID   //当前节点所在的集群ID raft cluster ID for request validation
	//Raft是一个接口，其实现的底层封装了etcd-raft模块，当rafthttp.Transport收到消息之后，会将其交给raft实例进行处理。
	Raft        Raft       // raft state machine, to which the Transport forwards received messages and reports status
	Snapshotter *snap.Snapshotter//负责管理快照文件
	ServerStats *stats.ServerStats // used to record general transportation statistics
	// used to record transportation statistics with followers when
	// performing as leader in raft protocol
	LeaderStats *stats.LeaderStats
	// ErrorC is used to report detected critical errors, e.g.,
	// the member has been permanently removed from the cluster
	// When an error is received from ErrorC, user should stop raft state
	// machine and thus stop the Transport.
	ErrorC chan error
	//stream消息通道中使用http.RoundTripper,默认是现实http.Transport,用来完成一个HTTP事务
	streamRt   http.RoundTripper // roundTripper used by streams
	//pipeline消息通道中使用http.RoundTripper,默认是现实http.Transport,用来完成一个HTTP事务
	pipelineRt http.RoundTripper // roundTripper used by pipelines

	mu      sync.RWMutex         // protect the remote and peer map
	//remotes中只封装了pipeline实例，remote主要负责发送快照数据，帮助新加入的节点快速追赶其他节点的数据。
	remotes map[types.ID]*remote // remotes map that helps newly joined member to catch up
	//Peer接口是当前节点对集群中其他节点的抽象表示。对于当前节点来说，集群中其他节点在本地都会有一个Peer实例与之对应，peers字段维护了节点ID到对应Peer实例之间的映射关系。
	peers   map[types.ID]Peer    // peers map
	//用于探测pipeline通道是否可用
	pipelineProber probing.Prober
	//用于探测stream通道是否可用
	streamProber   probing.Prober
}

func (t *Transport) Start() error {
	var err error
	//创建Stream消息通道使用http.RoundTripper实例，底层实际上是创建http.Transport实例
	//这里创建连接的超时时间(根据配置指定),读写请求的超时时间(默认5s),keepAlive时间(默认30s)
	t.streamRt, err = newStreamRoundTripper(t.TLSInfo, t.DialTimeout)
	if err != nil {
		return err
	}
	//创建Pipeline消息通道使用http.RoundTripper实例，底层实际上是创建http.Transport实例
	//创建连接的超时时间(根据配置指定),keepAlive时间(默认30s),与stream不同的是读写请求的超时时间设置为永不过期
	t.pipelineRt, err = NewRoundTripper(t.TLSInfo, t.DialTimeout)
	if err != nil {
		return err
	}
	//初始化remotes和peers两个map字段
	t.remotes = make(map[types.ID]*remote)
	t.peers = make(map[types.ID]Peer)
	//初始化prober实例用于探测pipeline和stream消息是否可用
	t.pipelineProber = probing.NewProber(t.pipelineRt)
	t.streamProber = probing.NewProber(t.streamRt)

	// If client didn't provide dial retry frequency, use the default
	// (100ms backoff between attempts to create a new stream),
	// so it doesn't bring too much overhead when retry.
	if t.DialRetryFrequency == 0 {
		t.DialRetryFrequency = rate.Every(100 * time.Millisecond)//拨号频次限定间隔100毫秒
	}
	return nil
}

func (t *Transport) Handler() http.Handler {
	/*
	 *三大Handler模块
	 *PipelineHandler
	 *StreamHandler
	 *SnapHandler
	 */
	pipelineHandler := newPipelineHandler(t, t.Raft, t.ClusterID)
	streamHandler := newStreamHandler(t, t, t.Raft, t.ID, t.ClusterID)
	snapHandler := newSnapshotHandler(t, t.Raft, t.Snapshotter, t.ClusterID)
	mux := http.NewServeMux()
	mux.Handle(RaftPrefix, pipelineHandler)
	mux.Handle(RaftStreamPrefix+"/", streamHandler)
	mux.Handle(RaftSnapshotPrefix, snapHandler)
	mux.Handle(ProbingPrefix, probing.NewHandler())
	return mux
}

func (t *Transport) Get(id types.ID) Peer {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.peers[id]
}

//发送消息
func (t *Transport) Send(msgs []raftpb.Message) {
	for _, m := range msgs {//遍历msgs切片中的全部消息
		if m.To == 0 {
			// ignore intentionally dropped message
			continue
		}
		//根据raftpb.Message.To字段,获取目标节点对应的Peer实例
		to := types.ID(m.To)

		t.mu.RLock()
		p, pok := t.peers[to]
		g, rok := t.remotes[to]
		t.mu.RUnlock()

		if pok {//如果存在对应的Peer实例，则使用Peer发送消息
			if m.Type == raftpb.MsgApp {
				t.ServerStats.SendAppendReq(m.Size())
			}
			p.send(m)//通过peer.send()方法完成消息的发送
			continue
		}

		if rok {//如果指定节点ID不存在对应的Peer实例，则尝试使用查找对应的remote实例
			g.send(m)//通过remote.send()方法完成消息的发送
			continue
		}

		if t.Logger != nil {
			t.Logger.Debug(
				"ignored message send request; unknown remote peer target",
				zap.String("type", m.Type.String()),
				zap.String("unknown-target-peer-id", to.String()),
			)
		} else {
			plog.Debugf("ignored message %s (sent to unknown peer %s)", m.Type, to)
		}
	}
}

func (t *Transport) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, r := range t.remotes {
		r.stop()
	}
	for _, p := range t.peers {
		p.stop()
	}
	t.pipelineProber.RemoveAll()
	t.streamProber.RemoveAll()
	if tr, ok := t.streamRt.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}
	if tr, ok := t.pipelineRt.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}
	t.peers = nil
	t.remotes = nil
}

// CutPeer drops messages to the specified peer.
func (t *Transport) CutPeer(id types.ID) {
	t.mu.RLock()
	p, pok := t.peers[id]
	g, gok := t.remotes[id]
	t.mu.RUnlock()

	if pok {
		p.(Pausable).Pause()
	}
	if gok {
		g.Pause()
	}
}

// MendPeer recovers the message dropping behavior of the given peer.
func (t *Transport) MendPeer(id types.ID) {
	t.mu.RLock()
	p, pok := t.peers[id]
	g, gok := t.remotes[id]
	t.mu.RUnlock()

	if pok {
		p.(Pausable).Resume()
	}
	if gok {
		g.Resume()
	}
}

func (t *Transport) AddRemote(id types.ID, us []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.remotes == nil {
		// there's no clean way to shutdown the golang http server
		// (see: https://github.com/golang/go/issues/4674) before
		// stopping the transport; ignore any new connections.
		return
	}
	if _, ok := t.peers[id]; ok {
		return
	}
	if _, ok := t.remotes[id]; ok {
		return
	}
	urls, err := types.NewURLs(us)
	if err != nil {
		if t.Logger != nil {
			t.Logger.Panic("failed NewURLs", zap.Strings("urls", us), zap.Error(err))
		} else {
			plog.Panicf("newURLs %+v should never fail: %+v", us, err)
		}
	}
	t.remotes[id] = startRemote(t, urls, id)

	if t.Logger != nil {
		t.Logger.Info(
			"added new remote peer",
			zap.String("local-member-id", t.ID.String()),
			zap.String("remote-peer-id", id.String()),
			zap.Strings("remote-peer-urls", us),
		)
	}
}

//AddPeer添加Peer
func (t *Transport) AddPeer(id types.ID, us []string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.peers == nil {
		panic("transport stopped")
	}
	if _, ok := t.peers[id]; ok {
		return
	}
	urls, err := types.NewURLs(us)//解析us切片中指定URL连接
	if err != nil {
		if t.Logger != nil {
			t.Logger.Panic("failed NewURLs", zap.Strings("urls", us), zap.Error(err))
		} else {
			plog.Panicf("newURLs %+v should never fail: %+v", us, err)
		}
	}
	fs := t.LeaderStats.Follower(id.String())
	//创建指定节点对应的Peer实例，其中会相关的Stream消息通道和Pipeline消息通道
	t.peers[id] = startPeer(t, urls, id, fs)
	//每隔一段时间，prober会向该节点发送探测消息，检查对端的健康状况
	addPeerToProber(t.Logger, t.pipelineProber, id.String(), us, RoundTripperNameSnapshot, rttSec)
	addPeerToProber(t.Logger, t.streamProber, id.String(), us, RoundTripperNameRaftMessage, rttSec)

	if t.Logger != nil {
		t.Logger.Info(
			"added remote peer",
			zap.String("local-member-id", t.ID.String()),
			zap.String("remote-peer-id", id.String()),
			zap.Strings("remote-peer-urls", us),
		)
	} else {
		//控制台 add perr输出
		plog.Infof("added peer %s", id)
	}
}

func (t *Transport) RemovePeer(id types.ID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.removePeer(id)
}

func (t *Transport) RemoveAllPeers() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for id := range t.peers {
		t.removePeer(id)
	}
}

// the caller of this function must have the peers mutex.
func (t *Transport) removePeer(id types.ID) {
	if peer, ok := t.peers[id]; ok {
		peer.stop()
	} else {
		if t.Logger != nil {
			t.Logger.Panic("unexpected removal of unknown remote peer", zap.String("remote-peer-id", id.String()))
		} else {
			plog.Panicf("unexpected removal of unknown peer '%d'", id)
		}
	}
	delete(t.peers, id)
	delete(t.LeaderStats.Followers, id.String())
	t.pipelineProber.Remove(id.String())
	t.streamProber.Remove(id.String())

	if t.Logger != nil {
		t.Logger.Info(
			"removed remote peer",
			zap.String("local-member-id", t.ID.String()),
			zap.String("removed-remote-peer-id", id.String()),
		)
	} else {
		plog.Infof("removed peer %s", id)
	}
}

func (t *Transport) UpdatePeer(id types.ID, us []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	// TODO: return error or just panic?
	if _, ok := t.peers[id]; !ok {
		return
	}
	urls, err := types.NewURLs(us)
	if err != nil {
		if t.Logger != nil {
			t.Logger.Panic("failed NewURLs", zap.Strings("urls", us), zap.Error(err))
		} else {
			plog.Panicf("newURLs %+v should never fail: %+v", us, err)
		}
	}
	t.peers[id].update(urls)

	t.pipelineProber.Remove(id.String())
	addPeerToProber(t.Logger, t.pipelineProber, id.String(), us, RoundTripperNameSnapshot, rttSec)
	t.streamProber.Remove(id.String())
	addPeerToProber(t.Logger, t.streamProber, id.String(), us, RoundTripperNameRaftMessage, rttSec)

	if t.Logger != nil {
		t.Logger.Info(
			"updated remote peer",
			zap.String("local-member-id", t.ID.String()),
			zap.String("updated-remote-peer-id", id.String()),
			zap.Strings("updated-remote-peer-urls", us),
		)
	} else {
		plog.Infof("updated peer %s", id)
	}
}

func (t *Transport) ActiveSince(id types.ID) time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if p, ok := t.peers[id]; ok {
		return p.activeSince()
	}
	return time.Time{}
}

func (t *Transport) SendSnapshot(m snap.Message) {
	t.mu.Lock()
	defer t.mu.Unlock()
	p := t.peers[types.ID(m.To)]
	if p == nil {
		m.CloseWithError(errMemberNotFound)
		return
	}
	p.sendSnap(m)
}

// Pausable is a testing interface for pausing transport traffic.
type Pausable interface {
	Pause()
	Resume()
}

func (t *Transport) Pause() {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, p := range t.peers {
		p.(Pausable).Pause()
	}
}

func (t *Transport) Resume() {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, p := range t.peers {
		p.(Pausable).Resume()
	}
}

// ActivePeers returns a channel that closes when an initial
// peer connection has been established. Use this to wait until the
// first peer connection becomes active.
func (t *Transport) ActivePeers() (cnt int) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, p := range t.peers {
		if !p.activeSince().IsZero() {
			cnt++
		}
	}
	return cnt
}
