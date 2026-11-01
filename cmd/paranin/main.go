// Command paranin is the composition root of the site: it builds every part and hands each what it needs.
//
//	paranin check content    loads a content folder as the site would and lists every fault in it
//	paranin serve content    serves the site from a content folder over HTTP
//
// check prints the faults, one a line. serve listens at PARANIN_HTTP_ADDR (localhost:8080 when unset), loads the
// folder again on SIGHUP, keeping what it serves when the new content has faults, and stops on SIGINT or SIGTERM
// once the requests in hand are answered. The program exits 0 when it did what it was asked, 1 when the content
// has faults or the server failed, and 2 when the command itself is written wrong.
package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/hosting/web"
)

// What the program exits with
const (
	succeeded     = 0
	failed        = 1
	usageMistaken = 2
)

// How the command is written: the program, the command, the folder
const (
	checkCommand    = "check"
	serveCommand    = "serve"
	usage           = "usage: paranin check|serve <content folder>"
	commandArgument = 1
	folderArgument  = 2
	argumentCount   = 3
)

// Where the site is served, unless the environment says otherwise, and over what
const (
	addressVariable = "PARANIN_HTTP_ADDR"
	defaultAddress  = "localhost:8080"
	network         = "tcp"
)

// How faults are printed: the folder, quoted, then each fault on a line of its own
const faultsFormat = "%q:\n%v\n"

// The keys of what the log says
const (
	addressKey = "address"
	contentKey = "content"
	errorKey   = "error"
	faultsKey  = "faults"
)

// How long the server waits on a slow reader, and on the requests in hand when it stops
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 2 * time.Minute
	shutdownTimeout   = 30 * time.Second
)

func main() {
	os.Exit(run(os.Args))
}

func run(arguments []string) int {
	if len(arguments) != argumentCount {
		fmt.Fprintln(os.Stderr, usage)
		return usageMistaken
	}
	folderName := arguments[folderArgument]
	folder := os.DirFS(folderName)
	switch arguments[commandArgument] {
	case checkCommand:
		return check(folderName, folder)
	case serveCommand:
		return serve(folderName, folder)
	default:
		fmt.Fprintln(os.Stderr, usage)
		return usageMistaken
	}
}

func check(folderName string, folder fs.FS) int {
	if _, err := content.Load(folder); err != nil {
		fmt.Fprintf(os.Stderr, faultsFormat, folderName, err)
		return failed
	}
	return succeeded
}

func serve(folderName string, folder fs.FS) int {
	store, err := content.Open(folder)
	if err != nil {
		fmt.Fprintf(os.Stderr, faultsFormat, folderName, err)
		return failed
	}
	address := os.Getenv(addressVariable)
	if address == "" {
		address = defaultAddress
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	listener, err := net.Listen(network, address)
	if err != nil {
		logger.Error("the address cannot be listened at", addressKey, address, errorKey, err)
		return failed
	}
	stopping, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	running := serving{
		logger:      logger,
		store:       store,
		server:      newServer(web.New(store), logger),
		reloads:     make(chan os.Signal, 1),
		failures:    make(chan error, 1),
		stopSignals: stop,
	}
	signal.Notify(running.reloads, syscall.SIGHUP)
	go func() { running.failures <- running.server.Serve(listener) }()
	logger.Info("serving", addressKey, address, contentKey, folderName)
	return running.watch(stopping)
}

// newServer is a server of the handler that waits on no slow reader for long and logs its own errors too.
func newServer(handler http.Handler, logger *slog.Logger) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
}

// serving is the site while it is served: the store it reads, the server, and the signals it answers.
type serving struct {
	logger   *slog.Logger
	store    *content.Store
	server   *http.Server
	reloads  chan os.Signal
	failures chan error
	// stopSignals gives SIGINT and SIGTERM back, so a second one ends the program while the first one waits
	stopSignals context.CancelFunc
}

// watch answers each signal until the site stops: a reload, a stop, or the server failing.
func (running serving) watch(stopping context.Context) int {
	for {
		select {
		case <-stopping.Done():
			running.stopSignals()
			return running.shutDown(stopping)
		case <-running.reloads:
			running.reload()
		case err := <-running.failures:
			running.logger.Error("the server stopped", errorKey, err)
			return failed
		}
	}
}

// reload loads the content again; content with faults is refused, and the site serves what it served.
func (running serving) reload() {
	if err := running.store.Reload(); err != nil {
		running.logger.Error("the new content is refused", faultsKey, err.Error())
		return
	}
	running.logger.Info("the content is loaded again")
}

// shutDown stops the server once the requests in hand are answered, or the wait runs out. The stop has ended
// the context it came from, so the wait keeps that context's values alone.
func (running serving) shutDown(stopping context.Context) int {
	running.logger.Info("stopping")
	waiting, cancel := context.WithTimeout(context.WithoutCancel(stopping), shutdownTimeout)
	defer cancel()
	if err := running.server.Shutdown(waiting); err != nil {
		running.logger.Error("the server did not stop cleanly", errorKey, err)
		return failed
	}
	return succeeded
}
