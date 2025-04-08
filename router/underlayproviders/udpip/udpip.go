// Copyright 2025 SCION Association
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package udpip

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"maps"
	"net"
	"net/netip"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/scionproto/scion/pkg/addr"
	"github.com/scionproto/scion/pkg/log"
	"github.com/scionproto/scion/pkg/private/serrors"
	"github.com/scionproto/scion/pkg/slayers"
	"github.com/scionproto/scion/private/underlay/conn"
	"github.com/scionproto/scion/router"
	"github.com/scionproto/scion/router/bfd"
)

var (
	errResolveOnSiblingLink  = errors.New("unsupported address resolution on sibling link")
	errResolveOnExternalLink = errors.New("unsupported address resolution on external link")
	errInvalidServiceAddress = errors.New("invalid service address")
	errShortPacket           = errors.New("packet is too short")
	errDuplicateRemote       = errors.New("duplicate remote address")
)

// An interface to enable unit testing.
type ConnNewer interface {
	New(l netip.AddrPort, r netip.AddrPort, c *conn.Config) (router.BatchConn, error)
}

// The default ConnNewer for this underlay: opens an udp BatchConn.
type un struct {
}

func (_ un) New(l netip.AddrPort, r netip.AddrPort, c *conn.Config) (router.BatchConn, error) {
	return conn.New(l, r, c)
}

// provider implements UnderlayProvider by making and returning Udp/Ip links.
//
// This is currently the only implementation. The goal of splitting out this code from the router
// is to enable other implementations. However, as a first step, we continue assuming that the
// batchConn is given to us and is a UDP socket and that, in the case of externalLink, it is bound.
type provider struct {
	mu                 sync.Mutex // Prevents race between adding connections and Start/Stop.
	batchSize          int
	allLinks           map[netip.AddrPort]udpLink
	allConnections     []*udpConnection
	connNewer          ConnNewer // un{}, except for unit tests
	svc                *router.Services[netip.AddrPort]
	internalConnection *udpConnection // Because we share it w/ siblinglinks
	internalHashSeed   uint32         // As a result, this too is shared.
	receiveBufferSize  int
	sendBufferSize     int
	dispatchStart      uint16
	dispatchEnd        uint16
	dispatchRedirect   uint16
}

type udpLink interface {
	router.Link
	start(ctx context.Context, procQs []chan *router.Packet, pool router.PacketPool)
	stop()
	receive(size int, srcAddr *net.UDPAddr, p *router.Packet)
}

func init() {
	// Register ourselves as an underlay provider. The registration consists of a constructor, not
	// a provider object, because multiple router instances each must have their own underlay
	// provider. The provider is not re-entrant.
	router.AddUnderlay("udpip", newProvider)
}

// New instantiates a new instance of the provider for exclusive use by the caller.
// TODO(multi_underlay): batchSize should be an underlay-specific config.
func newProvider(batchSize int, receiveBufferSize int, sendBufferSize int) router.UnderlayProvider {
	return &provider{
		batchSize:         batchSize,
		allLinks:          make(map[netip.AddrPort]udpLink),
		connNewer:         un{},
		svc:               router.NewServices[netip.AddrPort](),
		receiveBufferSize: receiveBufferSize,
		sendBufferSize:    sendBufferSize,
	}
}

// SetConnNewer installs the given newer. newer must be an implementation of ConnNewer or
// panic will ensue. Only for use in unit tests.
func (u *provider) SetConnNewer(newer any) {
	u.connNewer = newer.(ConnNewer)
}

func (u *provider) NumConnections() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return len(u.allLinks)
}

func (u *provider) Headroom() int {
	// This underlay does not add any header of its own: the UDP socket API manages the header
	// independently.
	return 0
}

func (u *provider) SetDispatchPorts(start, end, redirect uint16) {
	u.dispatchStart = start
	u.dispatchEnd = end
	u.dispatchRedirect = redirect
}

// AddSvc adds the address for the given service.
func (u *provider) AddSvc(svc addr.SVC, a addr.Host, p uint16) error {
	// We pre-resolve the addresses, which is trivial for this underlay.
	addr := netip.AddrPortFrom(a.IP(), p)
	if !addr.IsValid() {
		return errInvalidServiceAddress
	}
	u.svc.AddSvc(svc, addr)
	return nil
}

// DelSvc deletes the address for the given service.
func (u *provider) DelSvc(svc addr.SVC, a addr.Host, p uint16) error {
	addr := netip.AddrPortFrom(a.IP(), p)
	if !addr.IsValid() {
		return errInvalidServiceAddress
	}
	u.svc.DelSvc(svc, addr)
	return nil
}

// The queues to be used by the receiver task are supplied at this point because they must be
// sized according to the number of connections that will be started.
func (u *provider) Start(
	ctx context.Context, pool router.PacketPool, procQs []chan *router.Packet,
) {
	u.mu.Lock()
	if len(procQs) == 0 {
		// Pointless to run without any processor of incoming traffic
		return
	}
	connSnapshot := slices.Clone(u.allConnections)
	linkSnapshot := slices.Collect(maps.Values(u.allLinks))
	u.mu.Unlock()

	// Links MUST be started before connections. Given that this is an internal mater, we don't pay
	// the price of checking at use time.
	for _, l := range linkSnapshot {
		l.start(ctx, procQs, pool)
	}
	for _, c := range connSnapshot {
		c.start(u.batchSize, pool)
	}
}

func (u *provider) Stop() {
	u.mu.Lock()
	connSnapshot := slices.Clone(u.allConnections)
	linkSnapshot := slices.Collect(maps.Values(u.allLinks))
	u.mu.Unlock()

	for _, c := range connSnapshot {
		c.stop()
	}
	for _, l := range linkSnapshot {
		l.stop()
	}
}

// udpConnection is essentially a BatchConn with a sending queue. The rest is about logs and
// metrics. This allows UDP connections to be shared between links. Bundling link and
// connection together is possible and simpler for the code here, but leaks more refactoring changes
// in the main router code. Specifically, either:
//   - sibling links would each need an independent socket to the sibling router, which
//     the router cannot provide at the moment.
//   - the internal links and sibling links would be the same, which means the router needs to
//     special case the sibling links: which we want to remove from the main code.
type udpConnection struct {
	conn         router.BatchConn
	queue        chan *router.Packet
	metrics      router.InterfaceMetrics
	name         string // for logs. It's more informative than ifID.
	running      atomic.Bool
	receiverDone chan struct{}
	senderDone   chan struct{}
	link         *externalLink              // External Link with exclusive use of the connection.
	links        map[netip.AddrPort]udpLink // Links that share this connection
}

// start puts the connection in the running state. In that state, the connection can deliver
// incoming packets and ignores packets present on its input channel.
func (u *udpConnection) start(batchSize int, pool router.PacketPool) {
	wasRunning := u.running.Swap(true)
	if wasRunning {
		return
	}

	// Receiver task
	go func() {
		defer log.HandlePanic()
		u.receive(batchSize, pool)
		close(u.receiverDone)
	}()

	// Forwarder task
	go func() {
		defer log.HandlePanic()
		u.send(batchSize, pool)
		close(u.senderDone)
	}()
}

// stop() puts the connection in the stopped state. In that state, the connection no longer delivers
// incoming packets and ignores packets present on its input channel. The connection is fully
// stopped when this method returns.
func (u *udpConnection) stop() {
	wasRunning := u.running.Swap(false)

	if wasRunning {
		u.conn.Close() // Unblock receiver
		close(u.queue) // Unblock sender
		<-u.receiverDone
		<-u.senderDone
	}
}

func (u *udpConnection) receive(batchSize int, pool router.PacketPool) {
	log.Debug("Receive", "connection", u.name)

	// A collection of socket messages, as the readBatch API expects them. We keep using the same
	// collection, call after call; only replacing the buffer.
	msgs := conn.NewReadMessages(batchSize)

	// An array of corresponding packet references. Each corresponds to one msg.
	// The packet owns the buffer that we set in the matching msg, plus the metadata that we'll add.
	packets := make([]*router.Packet, batchSize)
	numReusable := 0 // unused buffers from previous loop

	for u.running.Load() {
		// collect packets.

		// Give a new buffer to the msgs elements that have been used in the previous loop.
		for i := 0; i < batchSize-numReusable; i++ {
			p := pool.Get()
			packets[i] = p
			msgs[i].Buffers[0] = p.RawPacket
		}

		// Fill the packets
		numReusable = len(msgs)
		numPkts, err := u.conn.ReadBatch(msgs)
		if err != nil {
			log.Debug("Error while reading batch", "connection", u.name, "err", err)
			continue
		}
		numReusable -= numPkts
		for i, msg := range msgs[:numPkts] {

			// Update size; readBatch does not.
			size := msg.N
			p := packets[i]
			p.RawPacket = p.RawPacket[:size]

			// Find the right link. For unshared connections, it's easy: we know the link.
			// TODO(multi_underlay): this may justify creating multiple udpConnection
			// implementations?. For example, converting the srcAddr to a netip.AddrPort
			// is expensive; we could pass it to receive, but we wouldn't want to do it
			// for bound connections.
			if u.link != nil {
				u.link.receive(size, msg.Addr.(*net.UDPAddr), p)
				continue
			}

			// Ok then, find it by remote address. We have a map of *our* links, so it's short.
			srcAddr := msg.Addr.(*net.UDPAddr).AddrPort()
			l, found := u.links[srcAddr]
			if !found {
				// Anything else is the internal link.
				l = u.links[netip.AddrPort{}]
			}
			l.receive(size, msg.Addr.(*net.UDPAddr), p)
		}
	}

	// We have to stop receiving. Return the unsent packets to the pool to avoid creating
	// a memory leak (it is likely but not required that the process will exit).
	for _, p := range packets[batchSize-numReusable : batchSize] {
		pool.Put(p)
	}
}

func readUpTo(queue <-chan *router.Packet, n int, needsBlocking bool, pkts []*router.Packet) int {
	i := 0
	if needsBlocking {
		p, ok := <-queue
		if !ok {
			return i
		}
		pkts[i] = p
		i++
	}

	for ; i < n; i++ {
		select {
		case p, ok := <-queue:
			if !ok {
				return i
			}
			pkts[i] = p
		default:
			return i
		}
	}
	return i
}

// TODO(jiceatscion): There is a big issue with metrics and ifID. If an underlay connection must be
// shared between links (for example, sibling links), then we don't have a specific ifID in the
// connection per se. It changes for each packet. As a result, in the shared case, either we account
// all metrics to whatever placeholder ifID we have (i.e. 0), or we have to use pkt.egress and
// lookup the metrics in the map for each packet. This is too expensive.
//
// Mitigations:
//   - use ifID even if it is 0 for sibling links - no worse than before, since sibling links were
//     already redirected to interface 0 (...until we have fully shared forwarders - like with an
//     XDP underlay impl).
//   - stage our own internal metrics map, sorted by ifID = pkt.egress, and batch update the
//     metrics... might not be much cheaper than the naive way.
//   - Use one fw queue per ifID in each connection... but then have to round-robin for fairness....
//     smaller batches?
//
// For now, we do the first option. Whether that is good enough is still TBD.

func (u *udpConnection) send(batchSize int, pool router.PacketPool) {
	log.Debug("Send", "connection", u.name)

	// We use this somewhat like a ring buffer.
	pkts := make([]*router.Packet, batchSize)

	// We use this as a temporary buffer, but allocate it just once
	// to save on garbage handling.
	msgs := make(conn.Messages, batchSize)
	for i := range msgs {
		msgs[i].Buffers = make([][]byte, 1)
	}

	queue := u.queue
	conn := u.conn
	metrics := u.metrics
	toWrite := 0

	for u.running.Load() {
		// Top-up our batch.
		toWrite += readUpTo(queue, batchSize-toWrite, toWrite == 0, pkts[toWrite:])

		// Turn the packets into underlay messages that WriteBatch can send.
		for i, p := range pkts[:toWrite] {
			msgs[i].Buffers[0] = p.RawPacket
			msgs[i].Addr = nil
			// If we're a bound connection we do not need to specify the address. In fact
			// we really should not as it can cause unnecessary route queries. If we're
			// an unbound connection, we *must* specify the address, of course.
			if u.link == nil {
				msgs[i].Addr = p.RemoteAddr
			}
		}

		written, _ := conn.WriteBatch(msgs[:toWrite], 0)
		if written < 0 {
			// WriteBatch returns -1 on error, we just consider this as
			// 0 packets written
			written = 0
		}
		router.UpdateOutputMetrics(metrics, pkts[:written])
		for _, p := range pkts[:written] {
			pool.Put(p)
		}
		if written != toWrite {
			// Only one is dropped at this time. We'll retry the rest.
			sc := router.ClassOfSize(len(pkts[written].RawPacket))
			metrics[sc].DroppedPacketsInvalid.Inc()
			pool.Put(pkts[written])
			toWrite -= (written + 1)
			// Shift the leftovers to the head of the buffers.
			for i := 0; i < toWrite; i++ {
				pkts[i] = pkts[i+written+1]
			}
		} else {
			toWrite = 0
		}
	}
}

// makeHashSeed creates a new random number to serve as hash seed.
// Each receive loop is associated with its own hash seed to compute
// the proc queue where a packet should be delivered. All links that share
// an underlying connection (therefore a receive loop) use the same hash seed.
func makeHashSeed() uint32 {
	hashSeed := fnv1aOffset32
	randomBytes := make([]byte, 4)
	if _, err := rand.Read(randomBytes); err != nil {
		panic("Error while generating random value")
	}
	for _, c := range randomBytes {
		hashSeed = hashFNV1a(hashSeed, c)
	}
	return hashSeed
}

// TODO(jiceatscion): use more inheritance between implementations?

type externalLink struct {
	procQs     []chan *router.Packet
	egressQ    chan<- *router.Packet
	metrics    router.InterfaceMetrics
	pool       router.PacketPool
	bfdSession *bfd.Session
	seed       uint32
	ifID       uint16
}

// NewExternalLink returns an external link over the UdpIpUnderlay.
//
// TODO(multi_underlay): we get the connection ready-made and require it to be bound. So, we
// don't keep the remote address, but in the future, we will be making the connections (and
// the conn argument will be gone), so we will need the address for that.
func (u *provider) NewExternalLink(
	qSize int,
	bfd *bfd.Session,
	local string,
	remote string,
	ifID uint16,
	metrics router.InterfaceMetrics,
) (router.Link, error) {

	localAddr, err := conn.ResolveAddrPortOrPort(local)
	if err != nil {
		return nil, serrors.Wrap("resolving local address", err)
	}
	remoteAddr, err := conn.ResolveAddrPort(remote)
	if err != nil {
		return nil, serrors.Wrap("resolving remote address", err)
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	if l := u.allLinks[remoteAddr]; l != nil {
		// This is a really bad idea. We can't just panic because it may be a configuration error.
		// So, we have to return an error.
		return nil, serrors.Join(errDuplicateRemote, nil, "addr", remote)
	}

	conn, err := u.connNewer.New(localAddr, remoteAddr,
		&conn.Config{ReceiveBufferSize: u.receiveBufferSize, SendBufferSize: u.sendBufferSize})
	if err != nil {
		return nil, err
	}

	queue := make(chan *router.Packet, qSize)
	el := &externalLink{
		egressQ:    queue,
		metrics:    metrics,
		ifID:       ifID,
		bfdSession: bfd,
		seed:       makeHashSeed(),
	}
	c := &udpConnection{
		conn:         conn,
		queue:        queue,
		metrics:      metrics, // send() needs them :-(
		name:         remote,
		receiverDone: make(chan struct{}),
		senderDone:   make(chan struct{}),
		link:         el,
	}
	u.allConnections = append(u.allConnections, c)
	u.allLinks[remoteAddr] = el
	return el, nil
}

func (l *externalLink) start(
	ctx context.Context,
	procQs []chan *router.Packet,
	pool router.PacketPool,
) {
	// procQs and pool are never known before all configured links have been instantiated.  So we
	// get them only now. We didn't need it earlier since the connections have not been started yet.
	l.procQs = procQs
	l.pool = pool
	if l.bfdSession == nil {
		return
	}
	go func() {
		defer log.HandlePanic()
		if err := l.bfdSession.Run(ctx); err != nil && !errors.Is(err, bfd.AlreadyRunning) {
			log.Error("BFD session failed to start", "external interface", l.ifID, "err", err)
		}
	}()
}

func (l *externalLink) stop() {
	if l.bfdSession == nil {
		return
	}
	l.bfdSession.Close()
}

func (l *externalLink) IfID() uint16 {
	return l.ifID
}

func (l *externalLink) Scope() router.LinkScope {
	return router.External
}

func (l *externalLink) BFDSession() *bfd.Session {
	return l.bfdSession
}

func (l *externalLink) IsUp() bool {
	return l.bfdSession == nil || l.bfdSession.IsUp()
}

// Resolve should not be useful on an external link so we don't implement it yet.
func (l *externalLink) Resolve(p *router.Packet, host addr.Host, port uint16) error {
	return errResolveOnExternalLink
}

func (l *externalLink) Send(p *router.Packet) bool {
	select {
	case l.egressQ <- p:
	default:
		return false
	}
	return true
}

func (l *externalLink) SendBlocking(p *router.Packet) {
	l.egressQ <- p
}

func (l *externalLink) receive(size int, srcAddr *net.UDPAddr, p *router.Packet) {
	metrics := l.metrics
	sc := router.ClassOfSize(size)
	metrics[sc].InputPacketsTotal.Inc()
	metrics[sc].InputBytesTotal.Add(float64(size))
	procID, err := computeProcID(p.RawPacket, len(l.procQs), l.seed)
	if err != nil {
		log.Debug("Error while computing procID", "err", err)
		l.pool.Put(p)
		metrics[sc].DroppedPacketsInvalid.Inc()
		return
	}

	p.Link = l
	// The src address does not need to be recorded in the packet. Even SCMP won't need it.
	select {
	case l.procQs[procID] <- p:
	default:
		l.pool.Put(p)
		metrics[sc].DroppedPacketsBusyProcessor.Inc()
	}
}

type siblingLink struct {
	procQs     []chan *router.Packet
	egressQ    chan<- *router.Packet
	metrics    router.InterfaceMetrics
	pool       router.PacketPool
	bfdSession *bfd.Session
	remote     *net.UDPAddr
	seed       uint32
}

// newSiblingLink returns a sibling link over the UdpIpUnderlay.
//
// TODO(multi_underlay): this can only be an improvement over internalLink if we have a bound
// batchConn with the sibling router. This is coming. Now the router gives us what we need
// to make the connection ourselves, but we still reuse the internal one as a transition.
// The router gives us a BFDSession in all cases and we might throw it away (there
// are no permanent resources attached to it). This could be fixed by moving some BFD related code
// in-here.
func (u *provider) NewSiblingLink(
	qSize int,
	bfd *bfd.Session,
	local string,
	remote string,
	metrics router.InterfaceMetrics,
) (router.Link, error) {

	// We don't currently use the address, but that will change. It *must* be valid.
	_, err := conn.ResolveAddrPortOrPort(local)
	if err != nil {
		return nil, serrors.Wrap("resolving local address", err)
	}
	remoteAddr, err := conn.ResolveAddrPort(remote)
	if err != nil {
		return nil, serrors.Wrap("resolving remote address", err)
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	// We silently deduplicate sibling links, so the router doesn't need to be aware or keep track
	// of link sharing.
	if l := u.allLinks[remoteAddr]; l != nil {
		return l, nil
	}

	// All sibling links re-use the internal connection. This used to be a late binding (packets to
	// siblings would get routed through the internal interface at run-time). But now this binding
	// happens right now and it can't work if this is called before newInternalLink.
	c := u.internalConnection
	if c == nil {
		// The router isn't supposed to do this. This is an internal error.
		// TODO(multi_underlay): This will go away when we stop sharing the internal connection.
		panic("newSiblingLink called before newInternalLink")
	}

	sl := &siblingLink{
		egressQ:    c.queue, // And therefore we do not use qsize for now.
		metrics:    metrics,
		bfdSession: bfd,
		remote:     net.UDPAddrFromAddrPort(remoteAddr),
		seed:       u.internalHashSeed, // per connection, but used only by link.
	}
	c.links[remoteAddr] = sl
	u.allLinks[remoteAddr] = sl
	return sl, nil
}

func (l *siblingLink) start(
	ctx context.Context,
	procQs []chan *router.Packet,
	pool router.PacketPool,
) {
	// procQs and pool are never known before all configured links have been instantiated.  So we
	// get them only now. We didn't need it earlier since the connections have not been started yet.
	l.procQs = procQs
	l.pool = pool
	if l.bfdSession == nil {
		return
	}
	go func() {
		defer log.HandlePanic()
		if err := l.bfdSession.Run(ctx); err != nil && !errors.Is(err, bfd.AlreadyRunning) {
			log.Error("BFD session failed to start", "remote address", l.remote, "err", err)
		}
	}()
}

func (l *siblingLink) stop() {
	if l.bfdSession == nil {
		return
	}
	l.bfdSession.Close()
}

func (l *siblingLink) IfID() uint16 {
	return 0
}

func (l *siblingLink) Scope() router.LinkScope {
	return router.Sibling
}

func (l *siblingLink) BFDSession() *bfd.Session {
	return l.bfdSession
}

func (l *siblingLink) IsUp() bool {
	return l.bfdSession == nil || l.bfdSession.IsUp()
}

// Resolve should not be useful on a sibling link so we don't implement it yet.
func (l *siblingLink) Resolve(p *router.Packet, host addr.Host, port uint16) error {
	return errResolveOnSiblingLink
}

func (l *siblingLink) Send(p *router.Packet) bool {
	// We use an unbound connection but we offer a connection-oriented service. So, we need to
	// supply the packet's destination address.
	updateRemoteAddr(p, l.remote)
	select {
	case l.egressQ <- p:
	default:
		return false
	}
	return true
}

func (l *siblingLink) SendBlocking(p *router.Packet) {
	// We use an unbound connection but we offer a connection-oriented service. So, we need to
	// supply the packet's destination address.
	updateRemoteAddr(p, l.remote)
	l.egressQ <- p
}

func (l *siblingLink) receive(size int, srcAddr *net.UDPAddr, p *router.Packet) {
	metrics := l.metrics
	sc := router.ClassOfSize(size)
	metrics[sc].InputPacketsTotal.Inc()
	metrics[sc].InputBytesTotal.Add(float64(size))
	procID, err := computeProcID(p.RawPacket, len(l.procQs), l.seed)
	if err != nil {
		log.Debug("Error while computing procID", "err", err)
		l.pool.Put(p)
		metrics[sc].DroppedPacketsInvalid.Inc()
		return
	}

	p.Link = l
	select {
	case l.procQs[procID] <- p:
	default:
		l.pool.Put(p)
		metrics[sc].DroppedPacketsBusyProcessor.Inc()
	}
}

type internalLink struct {
	procQs           []chan *router.Packet
	egressQ          chan *router.Packet
	metrics          router.InterfaceMetrics
	pool             router.PacketPool
	svc              *router.Services[netip.AddrPort]
	seed             uint32
	dispatchStart    uint16
	dispatchEnd      uint16
	dispatchRedirect uint16
}

// NewInternalLink returns a internal link over the UdpIpUnderlay.
//
// TODO(multi_underlay): We still go with the assumption that internal links are always
// udpip, so we don't expect a string here. That should change.
func (u *provider) NewInternalLink(
	localAddr netip.AddrPort, qSize int, metrics router.InterfaceMetrics) (router.Link, error) {

	u.mu.Lock()
	defer u.mu.Unlock()

	if u.internalConnection != nil {
		// We don't want to support this and the router doesn't do it. This is an internal error.
		panic("More than one internal link")
	}

	conn, err := u.connNewer.New(
		localAddr, netip.AddrPort{},
		&conn.Config{ReceiveBufferSize: u.receiveBufferSize, SendBufferSize: u.sendBufferSize})

	if err != nil {
		return nil, err
	}
	u.internalHashSeed = makeHashSeed()
	queue := make(chan *router.Packet, qSize)
	il := &internalLink{
		egressQ:          queue,
		metrics:          metrics,
		svc:              u.svc,
		seed:             u.internalHashSeed,
		dispatchStart:    u.dispatchStart,
		dispatchEnd:      u.dispatchEnd,
		dispatchRedirect: u.dispatchRedirect,
	}
	c := &udpConnection{
		conn:         conn,
		queue:        queue,
		metrics:      metrics, // send() needs them :-(
		name:         "internal",
		receiverDone: make(chan struct{}),
		senderDone:   make(chan struct{}),
		links:        make(map[netip.AddrPort]udpLink),
	}
	c.links[netip.AddrPort{}] = il
	u.allLinks[netip.AddrPort{}] = il
	u.internalConnection = c
	u.allConnections = append(u.allConnections, c)
	return il, nil
}

func (l *internalLink) start(
	ctx context.Context,
	procQs []chan *router.Packet,
	pool router.PacketPool,
) {
	// procQs and pool are never known before all configured links have been instantiated. So we
	// get them only now. We didn't need it earlier since the connections have not been started yet.
	l.procQs = procQs
	l.pool = pool
}

func (l *internalLink) stop() {
}

func (l *internalLink) IfID() uint16 {
	return 0
}

func (l *internalLink) Scope() router.LinkScope {
	return router.Internal
}

func (l *internalLink) BFDSession() *bfd.Session {
	return nil
}

func (l *internalLink) IsUp() bool {
	return true
}

// Resolve updates the packet's underlay destination according to the given SCION host/service
// address and SCION port number.  On the udpip underlay, host addresses are bit-for-bit identical
// to underlay addresses. The port space is the same, except if the packet is redirected to the shim
// dispatcher.  TODO(jiceatscion): make cheaper; copy addr less.
func (l *internalLink) Resolve(p *router.Packet, dst addr.Host, port uint16) error {
	var dstAddr netip.Addr
	switch dst.Type() {
	case addr.HostTypeSVC:
		// For map lookup use the Base address, i.e. strip the multi cast information, because we
		// only register base addresses in the map.
		a, ok := l.svc.Any(dst.SVC().Base())
		if !ok {
			return router.NoSVCBackend
		}
		dstAddr = a.Addr()
		// Supplied port is irrelevant. Port is in svc record.
		port = a.Port()
	case addr.HostTypeIP:
		dstAddr = dst.IP()
		if dstAddr.Is4In6() {
			return router.UnsupportedV4MappedV6Address
		}
		if dstAddr.IsUnspecified() {
			return router.UnsupportedUnspecifiedAddress
		}
	default:
		panic(fmt.Sprintf("unexpected address type returned from DstAddr: %s", dst.Type()))
	}
	// if port is outside the configured port range we send to the fixed port.
	if port < l.dispatchStart && port > l.dispatchEnd {
		port = l.dispatchRedirect
	}
	updateRemoteAddrFromAddrAndPort(p.RemoteAddr, dstAddr, port)
	return nil
}

// The packet's destination is already in the packet's meta-data.
func (l *internalLink) Send(p *router.Packet) bool {
	select {
	case l.egressQ <- p:
	default:
		return false
	}
	return true
}

// The packet's destination is already in the packet's meta-data.
func (l *internalLink) SendBlocking(p *router.Packet) {
	l.egressQ <- p
}

func (l *internalLink) receive(size int, srcAddr *net.UDPAddr, p *router.Packet) {
	metrics := l.metrics
	sc := router.ClassOfSize(size)
	metrics[sc].InputPacketsTotal.Inc()
	metrics[sc].InputBytesTotal.Add(float64(size))
	procID, err := computeProcID(p.RawPacket, len(l.procQs), l.seed)
	if err != nil {
		log.Debug("Error while computing procID", "err", err)
		l.pool.Put(p)
		metrics[sc].DroppedPacketsInvalid.Inc()
		return
	}

	p.Link = l
	// This is an unbound link. We must record the src address in case the packet
	// is turned around by SCMP.
	updateRemoteAddr(p, srcAddr)
	select {
	case l.procQs[procID] <- p:
	default:
		l.pool.Put(p)
		metrics[sc].DroppedPacketsBusyProcessor.Inc()
	}
}

func computeProcID(data []byte, numProcRoutines int, hashSeed uint32) (uint32, error) {
	if len(data) < slayers.CmnHdrLen {
		return 0, errShortPacket
	}
	dstHostAddrLen := slayers.AddrType(data[9] >> 4 & 0xf).Length()
	srcHostAddrLen := slayers.AddrType(data[9] & 0xf).Length()
	addrHdrLen := 2*addr.IABytes + srcHostAddrLen + dstHostAddrLen
	if len(data) < slayers.CmnHdrLen+addrHdrLen {
		return 0, errShortPacket
	}

	s := hashSeed

	// inject the flowID
	s = hashFNV1a(s, data[1]&0xF) // The left 4 bits aren't part of the flowID.
	for _, c := range data[2:4] {
		s = hashFNV1a(s, c)
	}

	// Inject the src/dst addresses
	for _, c := range data[slayers.CmnHdrLen : slayers.CmnHdrLen+addrHdrLen] {
		s = hashFNV1a(s, c)
	}

	return s % uint32(numProcRoutines), nil
}

// updateRemoteAddr() copies newAddr into the packet's RemoteAddr while re-using the IP slice
// embedded in it. This is to avoid giving work to the GC. Nil IPs get converted into empty slices.
// The backing array isn't discarded. A packet's remoteAddr is pre-initialized by the router with
// enough capacity for either V6 or V4 addresses. In the future it should become a generic storage
// area independent from any specific address family.
func updateRemoteAddr(p *router.Packet, newAddr *net.UDPAddr) {
	rAddr := p.RemoteAddr
	rAddr.Port = newAddr.Port
	rAddr.Zone = newAddr.Zone
	rAddr.IP = rAddr.IP[0:len(newAddr.IP)]
	copy(rAddr.IP, newAddr.IP)
}

// updateNetAddrFromAddrAndPort() updates a net.UDPAddr address to be the same IP and port as
// the given netip.Addr and port.
//
// We handle netAddr so we don't make the GC work. The issue is the IP address slice
// that's in netAddr. The packet, along with its address and IP slice, gets copied from a channel
// into a local variable. Then after we modify it, all gets copied to the some other channel and
// eventually it gets copied back to the pool. If we replace the destAddr.IP at any point,
// the old backing array behind the destAddr.IP slice ends-up on the garbage pile. To prevent that,
// we update the IP address in-place (we make the length 0 to represent the 0 address).
func updateRemoteAddrFromAddrAndPort(netAddr *net.UDPAddr, addr netip.Addr, port uint16) {
	netAddr.Port = int(port)
	netAddr.Zone = addr.Zone()
	if addr.Is4() {
		outIpBytes := addr.As4()     // Must store explicitly in order to copy
		netAddr.IP = netAddr.IP[0:4] // Update slice
		copy(netAddr.IP, outIpBytes[:])
	} else if addr.Is6() {
		outIpBytes := addr.As16()
		netAddr.IP = netAddr.IP[0:16]
		copy(netAddr.IP, outIpBytes[:])
	} else {
		// That's a zero address. We translate in to something resembling a nil IP.
		// Nothing gets discarded as we keep the slice (and its reference to the backing array).
		// To that end, we cannot make it nil. We have to make its length zero.
		netAddr.IP = netAddr.IP[0:0]
	}
}
