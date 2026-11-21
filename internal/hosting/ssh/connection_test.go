package ssh

import (
	"io"
	"net"
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
	client := &memoryConnection{incoming: toClient, outgoing: toServer, own: clientEnd, peer: serverEnd, unread: nil}
	server := &memoryConnection{incoming: toServer, outgoing: toClient, own: serverEnd, peer: clientEnd, unread: nil}
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
	unread []byte
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

// A connection in memory keeps no deadline: the bubble's clock and the test's end stand for it.
func (connection *memoryConnection) SetDeadline(time.Time) error { return nil }

func (connection *memoryConnection) SetReadDeadline(time.Time) error { return nil }

func (connection *memoryConnection) SetWriteDeadline(time.Time) error { return nil }

// memoryEndpoint is the address of either end of a connection in memory.
type memoryEndpoint struct{}

func (memoryEndpoint) Network() string { return memoryNetwork }

func (memoryEndpoint) String() string { return memoryAddress }
