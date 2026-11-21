package ssh

import (
	"io"
	"net"
	"os"
	"sync"
	"time"
)

// How many writes one end of a connection in memory holds before a write waits for the other end to read: enough
// for both ends to write their version line and their first packets at once, as SSH does
const chunksHeld = 64

// The network a connection in memory names, and its address
const (
	memoryNetwork = "memory"
	memoryAddress = "memory"
)

// memoryConnections are the client's end and the server's of a connection held in memory, so a test in a synctest
// bubble sees each goroutine of the server blocked as the bubble does, on a channel; a socket's are not.
func memoryConnections() (*memoryConnection, *memoryConnection) {
	toServer, toClient := make(chan []byte, chunksHeld), make(chan []byte, chunksHeld)
	clientEnd, serverEnd := newConnectionEnd(), newConnectionEnd()
	client := &memoryConnection{
		incoming: toClient, outgoing: toServer, own: clientEnd, peer: serverEnd, unread: nil,
		reading: newConnectionDeadline(), writing: newConnectionDeadline(),
	}
	server := &memoryConnection{
		incoming: toServer, outgoing: toClient, own: serverEnd, peer: clientEnd, unread: nil,
		reading: newConnectionDeadline(), writing: newConnectionDeadline(),
	}
	return client, server
}

// connectionEnd is one end of a connection in memory, which closes once.
type connectionEnd struct {
	closed  chan struct{}
	closing *sync.Once
}

func newConnectionEnd() connectionEnd {
	return connectionEnd{closed: make(chan struct{}), closing: new(sync.Once)}
}

// close closes the end, the first time it is called.
func (end connectionEnd) close() {
	end.closing.Do(func() { close(end.closed) })
}

// memoryConnection is one end of a connection in memory: each write is a chunk the other end reads.
type memoryConnection struct {
	incoming <-chan []byte
	outgoing chan<- []byte
	own      connectionEnd
	peer     connectionEnd
	// unread is what is left of the chunk read last
	unread  []byte
	reading *connectionDeadline
	writing *connectionDeadline
}

func (connection *memoryConnection) Read(buffer []byte) (int, error) {
	if len(connection.unread) == 0 {
		chunk, err := connection.nextChunk()
		if err != nil {
			return 0, err
		}
		connection.unread = chunk
	}
	count := copy(buffer, connection.unread)
	connection.unread = connection.unread[count:]
	return count, nil
}

// nextChunk is the next chunk the other end wrote, or the end of what it wrote once it has closed.
func (connection *memoryConnection) nextChunk() ([]byte, error) {
	select {
	case chunk := <-connection.incoming:
		return chunk, nil
	case <-connection.own.closed:
		return nil, net.ErrClosed
	case <-connection.peer.closed:
		return connection.lastChunk()
	case <-connection.reading.reached():
		return nil, os.ErrDeadlineExceeded
	}
}

// lastChunk is a chunk the other end wrote before it closed, or the end of what it wrote.
func (connection *memoryConnection) lastChunk() ([]byte, error) {
	select {
	case chunk := <-connection.incoming:
		return chunk, nil
	default:
		return nil, io.EOF
	}
}

func (connection *memoryConnection) Write(data []byte) (int, error) {
	select {
	case connection.outgoing <- append([]byte(nil), data...):
		return len(data), nil
	case <-connection.own.closed:
		return 0, net.ErrClosed
	case <-connection.peer.closed:
		return 0, net.ErrClosed
	case <-connection.writing.reached():
		return 0, os.ErrDeadlineExceeded
	}
}

func (connection *memoryConnection) Close() error {
	connection.own.close()
	return nil
}

func (connection *memoryConnection) LocalAddr() net.Addr {
	return memoryEndpoint{}
}

func (connection *memoryConnection) RemoteAddr() net.Addr {
	return memoryEndpoint{}
}

func (connection *memoryConnection) SetDeadline(moment time.Time) error {
	connection.reading.set(moment)
	connection.writing.set(moment)
	return nil
}

func (connection *memoryConnection) SetReadDeadline(moment time.Time) error {
	connection.reading.set(moment)
	return nil
}

func (connection *memoryConnection) SetWriteDeadline(moment time.Time) error {
	connection.writing.set(moment)
	return nil
}

// connectionDeadline is when a connection in memory gives up reading or writing, on the bubble's clock, as
// charm.land/ssh sets it for its timeouts; built as net.Pipe's, so a read or a write already waiting sees the
// deadline moved.
type connectionDeadline struct {
	mutex sync.Mutex
	timer *time.Timer
	// passed closes as the deadline passes, and is kept until it has: a fresh one stands only for one closed
	passed chan struct{}
}

func newConnectionDeadline() *connectionDeadline {
	return &connectionDeadline{mutex: sync.Mutex{}, timer: nil, passed: make(chan struct{})}
}

// set moves the deadline to a moment, or takes it away with the zero time.
func (deadline *connectionDeadline) set(moment time.Time) {
	deadline.mutex.Lock()
	defer deadline.mutex.Unlock()
	if deadline.timer != nil && !deadline.timer.Stop() {
		// the timer has fired: wait for it to close the channel
		<-deadline.passed
	}
	deadline.timer = nil
	if isClosed(deadline.passed) {
		deadline.passed = make(chan struct{})
	}
	if moment.IsZero() {
		return
	}
	if wait := time.Until(moment); wait > 0 {
		deadline.timer = time.AfterFunc(wait, func() { close(deadline.passed) })
		return
	}
	close(deadline.passed)
}

// reached closes as the deadline set now passes.
func (deadline *connectionDeadline) reached() <-chan struct{} {
	deadline.mutex.Lock()
	defer deadline.mutex.Unlock()
	return deadline.passed
}

// isClosed tells whether a channel is closed.
func isClosed(channel <-chan struct{}) bool {
	select {
	case <-channel:
		return true
	default:
		return false
	}
}

// memoryEndpoint is the address of either end of a connection in memory.
type memoryEndpoint struct{}

func (memoryEndpoint) Network() string { return memoryNetwork }

func (memoryEndpoint) String() string { return memoryAddress }
