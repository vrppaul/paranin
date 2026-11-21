// Command paranin is the composition root of the site: it builds every part and hands each what it needs.
//
//	paranin check content              loads a content folder as the site would and lists every fault in it
//	paranin check content build/web    loads a content folder and the page's build as one release, the same way
//	paranin serve content build/web    serves the site from a content folder and the page's build over HTTP and SSH
//
// check prints the faults, one a line: the commit's hook checks the content alone, and CI the release it builds.
// serve listens for HTTP at PARANIN_HTTP_ADDR (localhost:8080 when unset) and for SSH at PARANIN_SSH_ADDR
// (localhost:23234), with the host key in the file PARANIN_SSH_HOST_KEY names (.paranin/ssh_host_ed25519_key,
// made there when missing). It loads both folders again, as one release, on SIGHUP and once a change in either holds
// still, as `pnpm --dir web watch` writes the page again, keeping what it serves when either has faults, and stops
// on SIGINT or SIGTERM once the requests in hand are answered and the sessions open have ended, or the wait for
// them runs out. The program exits 0 when it did what it was asked, 1 when the content or the build has faults or
// the server failed, and 2 when the command itself is written wrong.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/hosting/ssh"
	"github.com/vrppaul/paranin/internal/hosting/web"
	"github.com/vrppaul/paranin/internal/release"
)

// What the program exits with
const (
	succeeded     = 0
	failed        = 1
	usageMistaken = 2
)

// How the command is written: the program, the command, the content folder, and the build folder a site is served
// with
const (
	checkCommand     = "check"
	serveCommand     = "serve"
	usage            = "usage: paranin check <content folder> [<build folder>] | paranin serve <content folder> <build folder>"
	commandArgument  = 1
	contentArgument  = 2
	buildArgument    = 3
	contentArguments = 3
	releaseArguments = 4
)

// Where the site is served, and the SSH server's host key, unless the environment says otherwise; and over what.
// The host key's file has sshd's name for one, in a folder only this program uses: wherever the server starts, no
// personal key is found there, and no folder of anyone else's is closed to 0700 when the key is made.
const (
	webAddressVariable = "PARANIN_HTTP_ADDR"
	defaultWebAddress  = "localhost:8080"
	sshAddressVariable = "PARANIN_SSH_ADDR"
	defaultSSHAddress  = "localhost:23234"
	hostKeyVariable    = "PARANIN_SSH_HOST_KEY"
	defaultHostKey     = ".paranin/ssh_host_ed25519_key"
	network            = "tcp"
)

// How faults are printed: the folder, or the folders of a release, each quoted, then each fault on a line of its own
const (
	faultsFormat    = "%s:\n%v\n"
	folderSeparator = " and "
)

// The keys of what the log says
const (
	webKey     = "http"
	sshKey     = "ssh"
	contentKey = "content"
	buildKey   = "build"
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

// How much of a request's head, its line and headers, the server reads; net/http answers 431 past it and its own
// 4 KB of slack. The bound is twice nginx's 8 KB for one header line, where Go's 1 MB lets one Accept cost about
// 100 ms to rank.
const maxHeaderBytes = 16 << 10

// How long an SSH session may stay idle and stay open (SEC-5): a reader who is only reading sends nothing, so
// the wait for a key is long
const (
	sessionIdleTimeout = 15 * time.Minute
	sessionMaxTimeout  = time.Hour
)

// How often the watch looks at the release's folders: it takes a change once the change has held still for a look,
// so a page rebuilt reaches the browser at most a second after the build's last write
const lookInterval = 500 * time.Millisecond

func main() {
	os.Exit(run(os.Args))
}

func run(arguments []string) int {
	switch command := commandOf(arguments); {
	case command == checkCommand && len(arguments) == contentArguments:
		return check(arguments[contentArgument])
	case command == checkCommand && len(arguments) == releaseArguments:
		return checkRelease(arguments[contentArgument], arguments[buildArgument])
	case command == serveCommand && len(arguments) == releaseArguments:
		return serve(arguments[contentArgument], arguments[buildArgument])
	default:
		fmt.Fprintln(os.Stderr, usage)
		return usageMistaken
	}
}

// commandOf is the command the arguments name, or nothing where they name none.
func commandOf(arguments []string) string {
	if len(arguments) <= commandArgument {
		return ""
	}
	return arguments[commandArgument]
}

func check(contentFolder string) int {
	_, err := content.Load(os.DirFS(contentFolder))
	return checked(err, contentFolder)
}

func checkRelease(contentFolder, buildFolder string) int {
	_, err := release.Load(release.FoldersAt(contentFolder, buildFolder))
	return checked(err, contentFolder, buildFolder)
}

// checked is what a check of folders exits with: success, or a failure once their faults are printed.
func checked(err error, folders ...string) int {
	if err != nil {
		return refused(err, folders...)
	}
	return succeeded
}

// refused prints the faults of folders that did not load, after the folders, quoted, and is what the program then
// exits with.
func refused(err error, folders ...string) int {
	quoted := make([]string, 0, len(folders))
	for _, folder := range folders {
		quoted = append(quoted, strconv.Quote(folder))
	}
	fmt.Fprintf(os.Stderr, faultsFormat, strings.Join(quoted, folderSeparator), err)
	return failed
}

func serve(contentFolder, buildFolder string) int {
	stopping, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	folders := release.FoldersAt(contentFolder, buildFolder)
	// The watch looks before the first load, so a change made while the site starts is a change it sees
	changes := release.Watch(stopping, folders, lookInterval)
	store, err := release.Open(folders)
	if err != nil {
		return refused(err, contentFolder, buildFolder)
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
	webServer := newServer(web.New(store), logger)
	servers := []server{webServer, sshServer}
	running := serving{
		logger:      logger,
		store:       store,
		servers:     servers,
		reloads:     make(chan os.Signal, 1),
		changes:     changes,
		failures:    make(chan error, len(servers)),
		stopSignals: stop,
	}
	signal.Notify(running.reloads, syscall.SIGHUP)
	go func() { running.failures <- webServer.Serve(webListener) }()
	go func() { running.failures <- sshServer.Serve(sshListener) }()
	logger.Info("serving", webKey, webAddress, sshKey, sshAddress, contentKey, contentFolder, buildKey, buildFolder)
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

// newServer is a server of the handler that waits on no slow reader for long, reads no head past its bound, and
// logs its own errors too.
func newServer(handler http.Handler, logger *slog.Logger) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}
}

// server is one way the site is served, which stops once what it has in hand is done or the wait runs out.
type server interface {
	Shutdown(ctx context.Context) error
}

// serving is the site while it is served: the store it reads, its servers, and the signals it answers.
type serving struct {
	logger  *slog.Logger
	store   *release.Store
	servers []server
	reloads chan os.Signal
	// changes says the folders changed and held still, as the release's watch saw them
	changes  <-chan struct{}
	failures chan error
	// stopSignals gives SIGINT and SIGTERM back, so a second one ends the program while the first one waits
	stopSignals context.CancelFunc
}

// watch answers each signal until the site stops: a reload, a change in the folders, a stop, or the server failing.
func (running serving) watch(stopping context.Context) int {
	for {
		select {
		case <-stopping.Done():
			running.stopSignals()
			return running.shutDown(stopping)
		case <-running.reloads:
			running.reload()
		case <-running.changes:
			running.reload()
		case err := <-running.failures:
			running.logger.Error("a server stopped", errorKey, err)
			return failed
		}
	}
}

// reload loads the release again; a release with faults in either folder is refused, and the site serves what it
// served.
func (running serving) reload() {
	if err := running.store.Reload(); err != nil {
		running.logger.Error("the new release is refused", faultsKey, err.Error())
		return
	}
	running.logger.Info("the release is loaded again")
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
