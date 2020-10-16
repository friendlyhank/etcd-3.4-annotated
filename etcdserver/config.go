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

package etcdserver

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/friendlyhank/etcd-3.4-annotated/pkg/netutil"
	"github.com/friendlyhank/etcd-3.4-annotated/pkg/transport"
	"github.com/friendlyhank/etcd-3.4-annotated/pkg/types"

	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ServerConfig holds the configuration of etcd as taken from the command line or discovery.
type ServerConfig struct {
	Name           string//etcdserver名称,对应flag "name"
	DiscoveryURL   string//etcd 用于发现服务,无需知道具体etcd节点ip即可访问etcd服务,对应flag "discovery"
	DiscoveryProxy string//供服务发现的代理地址,对应flag "discovery-proxy"
	// 由ip+port组成，默认DefaultListenClientURLs = "http://localhost:2379";
	// 实际情况使用https schema，用以外部etcd client访问，对应flag "listen-client-urls"
	ClientURLs     types.URLs
	// 由ip+port组成，默认DefaultListenPeerURLs   = "http://localhost:2380";
	// 实际生产环境使用http schema, 供etcd member 通信，对应flag "peer-client-urls"
	PeerURLs       types.URLs
	// 数据目录地址，为全路径，对应flag "data-dir"
	DataDir        string
	// DedicatedWALDir config will make the etcd to write the WAL to the WALDir
	// rather than the dataDir/member/wal.
	DedicatedWALDir string
	// 默认是10000次事件做一次快照:DefaultSnapCount = 100000
	// 可以作为调优参数进行参考，对应flag "snapshot-count",
	SnapshotCount uint64

	// SnapshotCatchUpEntries is the number of entries for a slow follower
	// to catch-up after compacting the raft storage entries.
	// We expect the follower has a millisecond level latency with the leader.
	// The max throughput is around 10K. Keep a 5K entries is enough for helping
	// follower to catch up.
	// WARNING: only change this for tests. Always use "DefaultSnapshotCatchUpEntries"
	SnapshotCatchUpEntries uint64
	// 默认是5，这是v2的参数，v3内只有一个db文件，
	// DefaultMaxSnapshots = 5，对应flag "max-snapshots"
	MaxSnapFiles uint
// 默认是5，DefaultMaxWALs      = 5，表示最大存储wal文件的个数，
	// 对应flag "max-wals"，保留的文件可以作为etcd-dump-logs工具进行debug使用。
	MaxWALFiles  uint

	// BackendBatchInterval is the maximum time before commit the backend transaction.
	BackendBatchInterval time.Duration
	// BackendBatchLimit is the maximum operations before commit the backend transaction.
	BackendBatchLimit int

	// BackendFreelistType is the type of the backend boltdb freelist.
	BackendFreelistType bolt.FreelistType
	// peerUrl与 etcd name对应的map,由方法cfg.PeerURLsMapAndToken("etcd")生成。
	InitialPeerURLsMap  types.URLsMap
	// etcd 集群token, 对应flang "initial-cluster-token"
	InitialClusterToken string
	// 确定是否为新建集群，对应flag "initial-cluster-state",
	// 由方法func (cfg Config) IsNewCluster() bool { return cfg.ClusterState == ClusterStateFlagNew }确定；
	NewCluster          bool
	PeerTLSInfo         transport.TLSInfo
	// member间通信使用的证书信息，若peerURL为https时使用，对应flag "peer-ca-file","peer-cert-file", "peer-key-file"
	CORS map[string]struct{}

	// HostWhitelist lists acceptable hostnames from client requests.
	// If server is insecure (no TLS), server only accepts requests
	// whose Host header value exists in this white list.
	HostWhitelist map[string]struct{}
	// raft node 发送心跳信息的超时时间。 "heartbeat-interval"
	TickMs        uint
	// raft node 发起选举的超时时间，最大为5000ms maxElectionMs = 50000,
	// 对应flag "election-timeout",
	// 选举时间与心跳时间在最佳实践内建议是10倍关系。
	ElectionTicks int

	// InitialElectionTickAdvance is true, then local member fast-forwards
	// election ticks to speed up "initial" leader election trigger. This
	// benefits the case of larger election ticks. For instance, cross
	// datacenter deployment may require longer election timeout of 10-second.
	// If true, local node does not need wait up to 10-second. Instead,
	// forwards its election ticks to 8-second, and have only 2-second left
	// before leader election.
	//
	// Major assumptions are that:
	//  - cluster has no active leader thus advancing ticks enables faster
	//    leader election, or
	//  - cluster already has an established leader, and rejoining follower
	//    is likely to receive heartbeats from the leader after tick advance
	//    and before election timeout.
	//
	// However, when network from leader to rejoining follower is congested,
	// and the follower does not receive leader heartbeat within left election
	// ticks, disruptive election has to happen thus affecting cluster
	// availabilities.
	//
	// Disabling this would slow down initial bootstrap process for cross
	// datacenter deployments. Make your own tradeoffs by configuring
	// --initial-election-tick-advance at the cost of slow initial bootstrap.
	//
	// If single-node, it advances ticks regardless.
	//
	// See https://github.com/etcd-io/etcd/issues/9333 for more detail.
	InitialElectionTickAdvance bool
	// etcd server启动的超时时间，默认为1s,
	// 由方法func (c *ServerConfig) bootstrapTimeout() time.Duration确定；
	BootstrapTimeout time.Duration
	// 默认为0，单位为小时，主要为了方便用户快速查询，
	// 定时对key进行合并处理，对应flag "auto-compaction-retention",
	// 由方法func NewPeriodic(h int, rg RevGetter, c Compactable) *Periodic确定，
	// 具体compact的实现方法为：
	//func (s *kvServer) Compact(ctx context.Context, r *pb.CompactionRequest) (*pb.CompactionResponse, error)
	AutoCompactionRetention time.Duration
	AutoCompactionMode      string
	// etcd后端数据文件的大小，默认为2GB，最大为8GB, v3的参数，
	// 对应flag  "quota-backend-bytes" ，具体定义：etcd\etcdserver\quota.go
	CompactionBatchLimit    int
	QuotaBackendBytes       int64
	MaxTxnOps               uint

	// MaxRequestBytes is the maximum request size to send over raft.
	//grpc最大请求bytes
	MaxRequestBytes uint

	StrictReconfigCheck bool

	// ClientCertAuthEnabled is true when cert has been signed by the client CA.
	ClientCertAuthEnabled bool

	AuthToken  string
	BcryptCost uint

	// InitialCorruptCheck is true to check data corruption on boot
	// before serving any peer/client traffic.
	InitialCorruptCheck bool
	CorruptCheckTime    time.Duration

	// PreVote is true to enable Raft Pre-Vote.
	PreVote bool//是否开启预候选人选举

	// Logger logs server-side operations.
	// If not nil, it disables "capnslog" and uses the given logger.
	Logger *zap.Logger

	// LoggerConfig is server logger configuration for Raft logger.
	// Must be either: "LoggerConfig != nil" or "LoggerCore != nil && LoggerWriteSyncer != nil".
	LoggerConfig *zap.Config
	// LoggerCore is "zapcore.Core" for raft logger.
	// Must be either: "LoggerConfig != nil" or "LoggerCore != nil && LoggerWriteSyncer != nil".
	LoggerCore        zapcore.Core
	LoggerWriteSyncer zapcore.WriteSyncer

	Debug bool
	// 对应flag "force-new-cluster",默认为false,若为true，
	// 在生产环境内，一般用于含v2数据的集群恢复，
	// 效果为以现有数据或者空数据新建一个单节点的etcd集群，
	// 如果存在数据，则会清楚数据内的元数据信息，并重建只包含该etcd的元数据信息。
	ForceNewCluster bool

	// EnableLeaseCheckpoint enables primary lessor to persist lease remainingTTL to prevent indefinite auto-renewal of long lived leases.
	EnableLeaseCheckpoint bool
	// LeaseCheckpointInterval time.Duration is the wait duration between lease checkpoints.
	LeaseCheckpointInterval time.Duration

	EnableGRPCGateway bool
}

// VerifyBootstrap sanity-checks the initial config for bootstrap case
// and returns an error for things that should never happen.
func (c *ServerConfig) VerifyBootstrap() error {
	if err := c.hasLocalMember(); err != nil {
		return err
	}
	if err := c.advertiseMatchesCluster(); err != nil {
		return err
	}
	if checkDuplicateURL(c.InitialPeerURLsMap) {
		return fmt.Errorf("initial cluster %s has duplicate url", c.InitialPeerURLsMap)
	}
	if c.InitialPeerURLsMap.String() == "" && c.DiscoveryURL == "" {
		return fmt.Errorf("initial cluster unset and no discovery URL found")
	}
	return nil
}

// VerifyJoinExisting sanity-checks the initial config for join existing cluster
// case and returns an error for things that should never happen.
func (c *ServerConfig) VerifyJoinExisting() error {
	// The member has announced its peer urls to the cluster before starting; no need to
	// set the configuration again.
	if err := c.hasLocalMember(); err != nil {
		return err
	}
	if checkDuplicateURL(c.InitialPeerURLsMap) {
		return fmt.Errorf("initial cluster %s has duplicate url", c.InitialPeerURLsMap)
	}
	if c.DiscoveryURL != "" {
		return fmt.Errorf("discovery URL should not be set when joining existing initial cluster")
	}
	return nil
}

// hasLocalMember checks that the cluster at least contains the local server.
func (c *ServerConfig) hasLocalMember() error {
	if urls := c.InitialPeerURLsMap[c.Name]; urls == nil {
		return fmt.Errorf("couldn't find local name %q in the initial cluster configuration", c.Name)
	}
	return nil
}

// advertiseMatchesCluster confirms peer URLs match those in the cluster peer list.
func (c *ServerConfig) advertiseMatchesCluster() error {
	urls, apurls := c.InitialPeerURLsMap[c.Name], c.PeerURLs.StringSlice()
	urls.Sort()
	sort.Strings(apurls)
	ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
	defer cancel()
	ok, err := netutil.URLStringsEqual(ctx, c.Logger, apurls, urls.StringSlice())
	if ok {
		return nil
	}

	initMap, apMap := make(map[string]struct{}), make(map[string]struct{})
	for _, url := range c.PeerURLs {
		apMap[url.String()] = struct{}{}
	}
	for _, url := range c.InitialPeerURLsMap[c.Name] {
		initMap[url.String()] = struct{}{}
	}

	missing := []string{}
	for url := range initMap {
		if _, ok := apMap[url]; !ok {
			missing = append(missing, url)
		}
	}
	if len(missing) > 0 {
		for i := range missing {
			missing[i] = c.Name + "=" + missing[i]
		}
		mstr := strings.Join(missing, ",")
		apStr := strings.Join(apurls, ",")
		return fmt.Errorf("--initial-cluster has %s but missing from --initial-advertise-peer-urls=%s (%v)", mstr, apStr, err)
	}

	for url := range apMap {
		if _, ok := initMap[url]; !ok {
			missing = append(missing, url)
		}
	}
	if len(missing) > 0 {
		mstr := strings.Join(missing, ",")
		umap := types.URLsMap(map[string]types.URLs{c.Name: c.PeerURLs})
		return fmt.Errorf("--initial-advertise-peer-urls has %s but missing from --initial-cluster=%s", mstr, umap.String())
	}

	// resolved URLs from "--initial-advertise-peer-urls" and "--initial-cluster" did not match or failed
	apStr := strings.Join(apurls, ",")
	umap := types.URLsMap(map[string]types.URLs{c.Name: c.PeerURLs})
	return fmt.Errorf("failed to resolve %s to match --initial-cluster=%s (%v)", apStr, umap.String(), err)
}

func (c *ServerConfig) MemberDir() string { return filepath.Join(c.DataDir, "member") }

func (c *ServerConfig) WALDir() string {
	if c.DedicatedWALDir != "" {
		return c.DedicatedWALDir
	}
	return filepath.Join(c.MemberDir(), "wal")
}

func (c *ServerConfig) SnapDir() string { return filepath.Join(c.MemberDir(), "snap") }

func (c *ServerConfig) ShouldDiscover() bool { return c.DiscoveryURL != "" }

// ReqTimeout returns timeout for request to finish.
func (c *ServerConfig) ReqTimeout() time.Duration {
	// 5s for queue waiting, computation and disk IO delay
	// + 2 * election timeout for possible leader election
	return 5*time.Second + 2*time.Duration(c.ElectionTicks*int(c.TickMs))*time.Millisecond
}

func (c *ServerConfig) electionTimeout() time.Duration {
	return time.Duration(c.ElectionTicks*int(c.TickMs)) * time.Millisecond
}

func (c *ServerConfig) peerDialTimeout() time.Duration {
	// 1s for queue wait and election timeout
	return time.Second + time.Duration(c.ElectionTicks*int(c.TickMs))*time.Millisecond
}

func checkDuplicateURL(urlsmap types.URLsMap) bool {
	um := make(map[string]bool)
	for _, urls := range urlsmap {
		for _, url := range urls {
			u := url.String()
			if um[u] {
				return true
			}
			um[u] = true
		}
	}
	return false
}

func (c *ServerConfig) bootstrapTimeout() time.Duration {
	if c.BootstrapTimeout != 0 {
		return c.BootstrapTimeout
	}
	return time.Second
}

func (c *ServerConfig) backendPath() string { return filepath.Join(c.SnapDir(), "db") }
