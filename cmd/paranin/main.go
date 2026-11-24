// Command paranin is the composition root of the site: it builds every part and hands each what it needs.
//
//	paranin check content              loads a content folder as the site would and lists every fault in it
//	paranin check content build/web    loads a content folder and the page's build as one release, the same way
//	paranin serve content build/web    serves the site from a content folder and the page's build over HTTP and SSH
//	paranin serve --drafts content build/web    the same, the drafts too, each marked, for their author to read
//	paranin preview-link content <slug>         prints the link a post is previewed at, for thirty days or --lasts
//	paranin revise content <slug>               takes a revision of a post out, its current version, beside it
//
// check prints the faults, one a line: the commit's hook checks the content alone, and CI the release it builds. Every
// load, check's and serve's, holds the content to no two of the site's pages sharing an address, and to no post's link
// inside the site leading to no page or no section of one, now or as each post dated later comes out, the drafts'
// among them, which the content alone cannot know, since routes writes the addresses and compose names the anchors.
// serve listens for HTTP at PARANIN_HTTP_ADDR (localhost:8080 when unset) and for SSH at PARANIN_SSH_ADDR
// (localhost:23234), with the host key in the file PARANIN_SSH_HOST_KEY names (.paranin/ssh_host_ed25519_key,
// made there when missing). It serves each preview at the link a key signs, the key in the file PARANIN_PREVIEW_KEY
// names, and none where it names none; preview-link signs a link with the same key, so the link it prints opens on a
// server holding it. serve loads both folders again, as one release, on SIGHUP and once a change in either holds
// still, as `pnpm --dir web watch` writes the page again, keeping what it serves when either has faults, and stops
// on SIGINT or SIGTERM once the requests in hand are answered and the sessions open have ended, or the wait for
// them runs out. revise writes the post's title, lead and text as they stand to revisions/<day>.en.md beside it, the
// day its current version came out, and prints the file's path: its author then rewrites the post and dates the new
// version by its updated:, which the content's check asks for. The program exits 0 when it did what it was asked, 1
// when the content or the build has faults, the server failed, or no link could be printed or revision taken, and 2
// when the command itself is written wrong.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/content"
	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/hosting/ssh"
	"github.com/vrppaul/paranin/internal/hosting/web"
	"github.com/vrppaul/paranin/internal/release"
	"github.com/vrppaul/paranin/internal/routes"
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
	checkCommand       = "check"
	serveCommand       = "serve"
	previewLinkCommand = "preview-link"
	reviseCommand      = "revise"
	usage              = "usage: paranin check <content folder> [<build folder>] | paranin serve [--drafts] <content folder> <build folder> | " +
		"paranin preview-link [--lasts 720h, at most 2160h] <content folder> <slug> | paranin revise <content folder> <slug>"
	commandArgument  = 1
	contentArgument  = 2
	buildArgument    = 3
	contentArguments = 3
	releaseArguments = 4
	// revise's post's slug follows the content folder where a release's build folder would
	reviseSlugArgument = 3
	reviseArguments    = 4
)

// Who may read and write what revise writes in the content folder, as git checks a file of the repository out: its
// owner reads and writes, anyone reads, and a folder's owner lists and enters it too
const (
	revisionMode        = 0o644
	revisionsFolderMode = 0o755
)

// How revise writes its revision's file: once, and only where none is there yet
const newFileOnly = os.O_WRONLY | os.O_CREATE | os.O_EXCL

// How preview-link is written after its command: the flag of how long the link lasts, thirty days where it is not
// asked, a span a reader is sent a draft to read in, and ninety at most, so no link that leaks lives for years; then
// the content folder and the post's slug, by their places among the words past the flags
const (
	lastsFlag       = "lasts"
	lastsUsage      = "how long the link opens the preview, as Go writes a duration, at most 2160h"
	defaultLinkLife = 30 * 24 * time.Hour
	longestLinkLife = 90 * 24 * time.Hour
	slugPlace       = 1
	linkArguments   = 2
)

// How serve is written after its command: the flag that serves the drafts too, then the content folder and the build
// folder, by their places among the words past the flags
const (
	draftsFlag      = "drafts"
	draftsUsage     = "serve the drafts too, marked as drafts, for their author to read"
	contentPlace    = 0
	buildPlace      = 1
	servedArguments = 2
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

// The file of the key preview links are signed with, which no default names: a key made where none is found would
// differ from the one on the owner's machine, and its links would open nothing
const previewKeyVariable = "PARANIN_PREVIEW_KEY"

// errNoPreviewKey is a preview link asked for where the environment names no key to sign it with.
var errNoPreviewKey = errors.New("no preview key: " + previewKeyVariable + " names none")

// How faults are printed: the folder, or the folders of a release, each quoted, then each fault on a line of its own
const (
	faultsFormat    = "%s:\n%v\n"
	folderSeparator = " and "
)

// The keys of what the log says
const (
	webKey      = "http"
	sshKey      = "ssh"
	contentKey  = "content"
	buildKey    = "build"
	draftsKey   = "drafts"
	previewsKey = "previews"
	errorKey    = "error"
	faultsKey   = "faults"
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

// How long an SSH connection may take to start its session from when it connects, and a session may stay idle and
// stay open (SEC-5): the one bound covers the handshake, about three and a half round trips, and starting the
// session, about one and a half more, so twenty-five seconds give up on a link whose round trip passes about five,
// and are all a connection that sends nothing, or starts nothing, holds a file for; a reader who is only reading
// sends nothing, so the wait for a key is long
const (
	openingTimeout     = 25 * time.Second
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
	case command == serveCommand:
		asked, isWritten := servingOf(arguments[contentArgument:])
		if !isWritten {
			return usageWrong()
		}
		return serve(asked)
	case command == previewLinkCommand:
		asked, isWritten := linkAskedOf(arguments[contentArgument:])
		if !isWritten {
			return usageWrong()
		}
		return previewLink(asked)
	case command == reviseCommand && len(arguments) == reviseArguments:
		return revise(arguments[contentArgument], arguments[reviseSlugArgument])
	default:
		return usageWrong()
	}
}

// usageWrong prints how the command is written, and is what the program exits with when it is written wrong.
func usageWrong() int {
	fmt.Fprintln(os.Stderr, usage)
	return usageMistaken
}

// servingAsked is what serve is asked to serve: the release's folders, and whether the drafts too.
type servingAsked struct {
	folders release.Folders
	drafts  bool
}

// servingOf reads what serve is asked from the words after its command, and whether they are written right.
func servingOf(words []string) (servingAsked, bool) {
	flags := flag.NewFlagSet(serveCommand, flag.ContinueOnError)
	// a mistake is told by the usage alone, as for any other command
	flags.SetOutput(io.Discard)
	drafts := flags.Bool(draftsFlag, false, draftsUsage)
	if err := flags.Parse(words); err != nil || flags.NArg() != servedArguments {
		return servingAsked{folders: release.Folders{Content: "", Build: ""}, drafts: false}, false
	}
	return servingAsked{folders: release.Folders{Content: flags.Arg(contentPlace), Build: flags.Arg(buildPlace)}, drafts: *drafts}, true
}

// commandOf is the command the arguments name, or nothing where they name none.
func commandOf(arguments []string) string {
	if len(arguments) <= commandArgument {
		return ""
	}
	return arguments[commandArgument]
}

// check loads a content folder as the site would, and holds what it loads to the site's own check, as a release's
// load does.
func check(contentFolder string) int {
	loaded, err := release.ReadFolder(contentFolder, content.Load)
	if err == nil {
		err = checkedNow(loaded)
	}
	return checked(err, contentFolder)
}

func checkRelease(contentFolder, buildFolder string) int {
	_, err := release.Load(release.Folders{Content: contentFolder, Build: buildFolder}, checkedNow)
	return checked(err, contentFolder, buildFolder)
}

// checkedNow holds content to the site's own check, from the moment it is loaded on.
//
//wrapper:reads the clock as each load is checked, which the check takes as a value
func checkedNow(loaded *content.Snapshot) error {
	return compose.Check(loaded, time.Now()) //nolint:wrapcheck // the content's faults, printed one a line as they come
}

// previewLink prints the link a post of a content folder is previewed at, signed with the key the environment names,
// as a server holding the same key serves it; a failure where it names none or the post is no preview now.
func previewLink(asked linkAsked) int {
	key, err := readPreviewKey()
	if err != nil {
		return refused(err, previewKeyVariable)
	}
	if key == nil {
		return refused(errNoPreviewKey, previewKeyVariable)
	}
	now := time.Now()
	link, err := previewLinkOf(asked.folder, asked.slug, key, now, now.Add(asked.lasts))
	if err != nil {
		return refused(err, asked.folder)
	}
	fmt.Println(link)
	return succeeded
}

// linkAsked is what preview-link is asked: the content folder, the slug of the post, and how long the link lasts.
type linkAsked struct {
	folder string
	slug   string
	lasts  time.Duration
}

// linkAskedOf reads what preview-link is asked from the words after its command, and whether they are written right:
// a lifetime, where one is asked, of some time and no longer than the longest.
func linkAskedOf(words []string) (linkAsked, bool) {
	flags := flag.NewFlagSet(previewLinkCommand, flag.ContinueOnError)
	// a mistake is told by the usage alone, as for any other command
	flags.SetOutput(io.Discard)
	lasts := flags.Duration(lastsFlag, defaultLinkLife, lastsUsage)
	if err := flags.Parse(words); err != nil || flags.NArg() != linkArguments || *lasts <= 0 || *lasts > longestLinkLife {
		return linkAsked{folder: "", slug: "", lasts: 0}, false
	}
	return linkAsked{folder: flags.Arg(contentPlace), slug: flags.Arg(slugPlace), lasts: *lasts}, true
}

// previewLinkOf is the link a post of a content folder is previewed at at a moment, whole, signed with a key until it
// lapses: a fault where the folder does not load, or the post is then no preview, a draft, out, or none at all, since
// its link would open nothing.
func previewLinkOf(contentFolder, slug string, key *web.PreviewKey, moment, lapses time.Time) (string, error) {
	loaded, err := release.ReadFolder(contentFolder, content.Load)
	if err != nil {
		return "", err //nolint:wrapcheck // the content's faults, printed one a line as they come
	}
	previews := compose.SiteOf(compose.Prepare(loaded), compose.Audience{At: moment, Drafts: false}).Previews
	found := slices.IndexFunc(previews, func(previewed compose.Preview) bool { return previewed.Post.Head.Page.Slug == slug })
	if found < 0 {
		return "", fmt.Errorf("%q is no post previewed now: a preview, or a post dated later, of the folder %q", slug, contentFolder)
	}
	site := routes.New(loaded.Profile.Site.BaseURL, loaded.Profile.Site.TagAuthority).Previewing(slug, key.Token(slug, lapses))
	return site.Address(previews[found].Post.Head.Page, routes.FormatChosen), nil
}

// revise takes a revision of a post of a content folder out now, and prints the path it wrote it at.
func revise(contentFolder, slug string) int {
	written, err := reviseAt(contentFolder, slug, time.Now())
	if err != nil {
		return refused(err, contentFolder)
	}
	fmt.Println(written)
	return succeeded
}

// reviseAt takes a revision of the current version of a post of a content folder out at a moment, beside the post
// (docs/FOUNDATION.md §5.4, Revisions): the path it wrote, in the folder; a fault where the folder does not load, as it
// does not once a version is taken until the post is dated by a later update, where the post is then no post out, a
// draft, a preview or none at all, whose version no reader read, its version came out that day or later, or its text is
// its newest revision's already.
func reviseAt(contentFolder, slug string, moment time.Time) (string, error) {
	loaded, err := release.ReadFolder(contentFolder, content.Load)
	if err != nil {
		return "", err //nolint:wrapcheck // the content's faults, printed one a line as they come
	}
	served := compose.SiteOf(compose.Prepare(loaded), compose.Audience{At: moment, Drafts: false}).Posts
	found := slices.IndexFunc(loaded.Posts, func(post content.Post) bool { return post.Slug == slug })
	out := slices.ContainsFunc(served, func(post document.Document) bool { return post.Head.Page.Slug == slug })
	if found < 0 || !out {
		return "", fmt.Errorf("%q is no post out now of the folder %q: a revision is taken of a version a reader read", slug, contentFolder)
	}
	taken, err := content.RevisionOf(loaded.Posts[found], moment)
	if err != nil {
		return "", fmt.Errorf("revise %q: %w", slug, err)
	}
	return taken.Path, writeRevision(contentFolder, taken)
}

// writeRevision writes a revision's file in a content folder, opened as a root, so no path leads out of it, making its
// revisions folder where it is missing; never over a file there.
func writeRevision(contentFolder string, taken content.RevisionFile) error {
	root, err := os.OpenRoot(contentFolder)
	if err != nil {
		return fmt.Errorf("open the content folder %q: %w", contentFolder, err)
	}
	return errors.Join(writeRevisionIn(root, taken), root.Close())
}

// writeRevisionIn writes a revision's file in a root, never over a file there, and takes it away where it could not
// be written whole. A reload that meets the file half written refuses it as it refuses the whole one, which the
// post's current version's day names, until the post is dated by a later update.
func writeRevisionIn(root *os.Root, taken content.RevisionFile) error {
	if err := root.MkdirAll(path.Dir(taken.Path), revisionsFolderMode); err != nil {
		return fmt.Errorf("make the revisions folder of %q: %w", taken.Path, err)
	}
	file, err := root.OpenFile(taken.Path, newFileOnly, revisionMode)
	if err != nil {
		return fmt.Errorf("make the revision %q: %w", taken.Path, err)
	}
	_, err = file.Write(taken.Text)
	if err = errors.Join(err, file.Close()); err != nil {
		return fmt.Errorf("write the revision %q: %w", taken.Path, errors.Join(err, root.Remove(taken.Path)))
	}
	return nil
}

// readPreviewKey reads the key preview links are signed with from the file the environment names; none where it names
// none, so a site serves no preview link and prints none.
func readPreviewKey() (*web.PreviewKey, error) {
	keyPath := os.Getenv(previewKeyVariable)
	if keyPath == "" {
		return nil, nil
	}
	return web.ReadPreviewKey(keyPath) //nolint:wrapcheck // the error names the key's file and what failed
}

// checked is what a check of folders exits with: success, or a failure once their faults are printed.
func checked(err error, folders ...string) int {
	if err != nil {
		return refused(err, folders...)
	}
	return succeeded
}

// refused prints the faults of what did not load, folders or the preview key's setting, after what it is, quoted, and
// is what the program then exits with.
func refused(err error, sources ...string) int {
	quoted := make([]string, 0, len(sources))
	for _, source := range sources {
		quoted = append(quoted, strconv.Quote(source))
	}
	fmt.Fprintf(os.Stderr, faultsFormat, strings.Join(quoted, folderSeparator), err)
	return failed
}

func serve(asked servingAsked) int {
	stopping, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	folders := asked.folders
	// The watch looks before the first load, so a change made while the site starts is a change it sees
	changes := release.Watch(stopping, folders, lookInterval)
	store, err := release.Open(folders, checkedNow)
	if err != nil {
		return refused(err, folders.Content, folders.Build)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	settings := ssh.Settings{
		HostKeyPath: fromEnvironment(hostKeyVariable, defaultHostKey), OpeningTimeout: openingTimeout,
		IdleTimeout: sessionIdleTimeout, MaxTimeout: sessionMaxTimeout, Drafts: asked.drafts,
	}
	sshServer, err := ssh.New(store, logger, settings)
	if err != nil {
		logger.Error("the SSH server cannot be made", errorKey, err)
		return failed
	}
	previewKey, err := readPreviewKey()
	if err != nil {
		logger.Error("the preview key cannot be read", errorKey, err)
		return failed
	}
	webAddress, sshAddress := fromEnvironment(webAddressVariable, defaultWebAddress), fromEnvironment(sshAddressVariable, defaultSSHAddress)
	webListener, sshListener, err := listenAt(webAddress, sshAddress)
	if err != nil {
		logger.Error("the addresses cannot be listened at", errorKey, err)
		return failed
	}
	webServer := newServer(web.New(store, asked.drafts, previewKey), logger)
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
	logger.Info("serving", webKey, webAddress, sshKey, sshAddress, contentKey, folders.Content, buildKey, folders.Build, draftsKey, asked.drafts,
		previewsKey, previewKey != nil)
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
