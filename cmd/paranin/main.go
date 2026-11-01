// Command paranin is the composition root of the site: it builds every part and hands each what it needs.
//
//	paranin check content    loads a content folder as the site would and lists every fault in it
//	paranin serve content    serves the site from a content folder over HTTP and SSH
//
// check prints the faults, one a line. serve listens for HTTP at PARANIN_HTTP_ADDR (localhost:8080 when unset) and
// for SSH at PARANIN_SSH_ADDR (localhost:23234), with the host key in the file PARANIN_SSH_HOST_KEY names
// (.ssh/id_ed25519, made there when missing). It loads the folder again on SIGHUP, keeping what it serves when the
// new content has faults, and stops on SIGINT or SIGTERM once the requests in hand are answered and the sessions
// open have ended, or the wait for them runs out. The program exits 0 when it did what it was asked, 1 when the content
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
	"github.com/vrppaul/paranin/internal/hosting/ssh"
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

// Where the site is served, and the SSH server's host key, unless the environment says otherwise; and over what
const (
	webAddressVariable = "PARANIN_HTTP_ADDR"
	defaultWebAddress  = "localhost:8080"
	sshAddressVariable = "PARANIN_SSH_ADDR"
	defaultSSHAddress  = "localhost:23234"
	hostKeyVariable    = "PARANIN_SSH_HOST_KEY"
	defaultHostKey     = ".ssh/id_ed25519"
	network            = "tcp"
)

// How faults are printed: the folder, quoted, then each fault on a line of its own
const faultsFormat = "%q:\n%v\n"

// The keys of what the log says
const (
	webKey     = "http"
	sshKey     = "ssh"
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

// How long an SSH session may stay idle and stay open (SEC-5): a reader who is only reading sends nothing, so
// the wait for a key is long
const (
	sessionIdleTimeout = 15 * time.Minute
	sessionMaxTimeout  = time.Hour
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
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	settings := ssh.Settings{HostKeyPath: fromEnvironment(hostKeyVariable, defaultHostKey), IdleTimeout: sessionIdleTimeout, MaxTimeout: sessionMaxTimeout}
	sshServer, err := ssh.New(store, logger, settings)
	if err != nil {
		logger.Error("the SSH server cannot be made", errorKey, err)
		return failed
	}
	webAddress, sshAddress := fromEnvironment(webAddressVariable, defaultWebAddress), fromEnvironment(sshAddressVariable, defaultSSHAddress)
	webListener, sshListener, err := listenAt(webAddress, sshAddress)
	if err != nil {
		logger.Error("the addresses cannot be listened at", errorKey, err)
		return failed
	}
	stopping, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	webServer := newServer(web.New(store), logger)
	servers := []server{webServer, sshServer}
	running := serving{
		logger:      logger,
		store:       store,
		servers:     servers,
		reloads:     make(chan os.Signal, 1),
		failures:    make(chan error, len(servers)),
		stopSignals: stop,
	}
	signal.Notify(running.reloads, syscall.SIGHUP)
	go func() { running.failures <- webServer.Serve(webListener) }()
	go func() { running.failures <- sshServer.Serve(sshListener) }()
	logger.Info("serving", webKey, webAddress, sshKey, sshAddress, contentKey, folderName)
	return running.watch(stopping)
}

// fromEnvironment is the value of an environment variable, or a fallback where it is unset or empty.
func fromEnvironment(variable, fallback string) string {
	if value := os.Getenv(variable); value != "" {
		return value
	}
	return fallback
}

// listenAt listens at the web's address and the SSH server's, both or neither, before either server says it is
// serving.
func listenAt(webAddress, sshAddress string) (webListener, sshListener net.Listener, err error) {
	webListener, err = net.Listen(network, webAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("listen for HTTP at %q: %w", webAddress, err)
	}
	sshListener, err = net.Listen(network, sshAddress)
	if err != nil {
		// The first listener is given up with the second; its close can fail only on a socket already gone
		_ = webListener.Close()
		return nil, nil, fmt.Errorf("listen for SSH at %q: %w", sshAddress, err)
	}
	return webListener, sshListener, nil
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

// server is one way the site is served, which stops once what it has in hand is done or the wait runs out.
type server interface {
	Shutdown(ctx context.Context) error
}

// serving is the site while it is served: the store it reads, its servers, and the signals it answers.
type serving struct {
	logger   *slog.Logger
	store    *content.Store
	servers  []server
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
			running.logger.Error("a server stopped", errorKey, err)
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

// shutDown stops every server at once, so none takes new work while another finishes its own, and waits for
// each to finish what it has in hand, or for the one wait for them all to run out. The stop has ended the context
// it came from, so the wait keeps that context's values alone.
func (running serving) shutDown(stopping context.Context) int {
	running.logger.Info("stopping")
	waiting, cancel := context.WithTimeout(context.WithoutCancel(stopping), shutdownTimeout)
	defer cancel()
	faults := make(chan error, len(running.servers))
	for _, stopped := range running.servers {
		go func() { faults <- stopped.Shutdown(waiting) }()
	}
	code := succeeded
	for range running.servers {
		if err := <-faults; err != nil {
			running.logger.Error("a server did not stop cleanly", errorKey, err)
			code = failed
		}
	}
	return code
}
