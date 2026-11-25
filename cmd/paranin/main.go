// Command paranin is the composition root of the site: it builds every part and hands each what it needs.
//
//	paranin check content                        loads a content folder as the site would and lists every fault in it
//	paranin check content build/web build/cards  loads it, the page's build and the cards as one release, the same way
//	paranin serve content build/web build/cards  serves the site from a release's folders over HTTP and SSH
//	paranin serve --drafts content ...           the same, the drafts and posts dated later too, for their author to
//	                                             read, a draft marked as one and a post dated later by its day
//	paranin cards content build/cards            draws each page's card into the cards' folder, with their manifest
//	paranin revise content <slug>                takes a revision of a post out, its current version, beside it
//
// cards takes --watch to draw them again each time the content changes, until stopped; revise takes --language and a
// language's tag, as ru, for a post's translation into it.
//
// check prints the faults, one a line, and where there are none what the content warns of, a translation older than its
// original, which refuses nothing; serve logs it at each load. The commit's hook checks the content alone, and CI the
// release it builds. Every load, check's and serve's, holds the content to no two of the site's pages sharing an
// address, and to no post's link inside the site leading to no page or no section of one, now or as each post dated
// later comes out, the drafts' among them, which the content alone cannot know, since routes writes the addresses and
// compose names the anchors; a release's load holds its cards to one for every page a preview shows, now or as a post
// dated later comes out. cards draws those cards, run in the repository's root, where it finds the faces they are set
// in, and takes away the cards no page shows any more; it refuses a folder that holds anything but cards.
// serve listens for HTTP at PARANIN_HTTP_ADDR (localhost:8080 when unset) and for SSH at PARANIN_SSH_ADDR
// (localhost:23234), with the host key in the file PARANIN_SSH_HOST_KEY names (.paranin/ssh_host_ed25519_key,
// made there when missing). Where PARANIN_INDEXNOW_KEY sets a key, of IndexNow's form, and the drafts are not served,
// serve serves the key's file and tells IndexNow each address to read again as a reload takes or a post dated later
// comes out. serve loads the three folders again, as one release, on SIGHUP and once a change in any holds still, as
// `pnpm --dir web watch` writes the page again and `cards --watch` the cards, keeping what it serves when any has
// faults, and stops on SIGINT or SIGTERM once the requests in hand are answered and the sessions open have ended, or
// the wait for them runs out. revise writes the post's title, lead and text as they stand to
// revisions/<day>.<language>.md beside it, the day its current version came out, dates the post as updated on the day
// it runs, and prints the file's path: its author then rewrites the post, which loads as its new version. The program
// exits 0 when it did what it was asked, 1 when the content, the build or the cards have faults, the IndexNow key is
// of another form, the server failed, no card could be drawn, or no revision could be taken, and 2 when the command
// itself is written wrong.
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
	"github.com/vrppaul/paranin/internal/hosting/indexnow"
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

// How the command is written: the program, the command, the content folder, and the build folder and the cards' folder
// a site is served with
const (
	checkCommand  = "check"
	serveCommand  = "serve"
	cardsCommand  = "cards"
	reviseCommand = "revise"
	usage         = "usage: paranin check <content folder> [<build folder> <cards folder>] | " +
		"paranin serve [--drafts] <content folder> <build folder> <cards folder> | paranin cards [--watch] <content folder> <cards folder> | " +
		"paranin revise [--language ru] <content folder> <slug>"
	commandArgument      = 1
	contentArgument      = 2
	buildArgument        = 3
	releaseCardsArgument = 4
	contentArguments     = 3
	releaseArguments     = 5
)

// How revise names a post's translation, by the flag of its language; the original, in English, where none is asked
const (
	languageFlag  = "language"
	languageUsage = "the language of the post's translation, by its tag, as ru; none for the original"
	// how a fault names a translation: its slug, then its language
	translationName = "%q in %q"
)

// Who may read and write what the program writes, a revision in the content folder and a card, as git checks a file of
// the repository out and a build writes one: its owner reads and writes, anyone reads, and a folder's owner lists and
// enters it too
const (
	writtenFileMode   = 0o644
	writtenFolderMode = 0o755
)

// How revise writes its revision's file: once, and only where none is there yet
const newFileOnly = os.O_WRONLY | os.O_CREATE | os.O_EXCL

// How revise is written after its command: the content folder and the post's slug, by their places among the words
// past the flags
const (
	slugPlace     = 1
	postArguments = 2
)

// How serve is written after its command: the flag that serves the drafts too, then the content folder, the build
// folder and the cards' folder, by their places among the words past the flags
const (
	draftsFlag       = "drafts"
	draftsUsage      = "serve the drafts too, marked as drafts, for their author to read"
	contentPlace     = 0
	buildPlace       = 1
	servedCardsPlace = 2
	servedArguments  = 3
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

// The IndexNow key the site is known by, none where the environment sets none, and where IndexNow is told what
// changed: the endpoint every search engine that reads it shares (indexnow.org)
const (
	indexNowKeyVariable = "PARANIN_INDEXNOW_KEY"
	indexNowEndpoint    = "https://api.indexnow.org/indexnow"
)

// How faults are printed: the folder, or the folders of a release, each quoted, then each fault on a line of its own;
// and how warnings are, after the content folder, quoted
const (
	faultsFormat    = "%s:\n%v\n"
	warningsFormat  = "%s warns, refusing nothing:\n%v\n"
	folderSeparator = " and "
)

// The keys of what the log says
const (
	webKey      = "http"
	sshKey      = "ssh"
	contentKey  = "content"
	buildKey    = "build"
	cardsKey    = "cards"
	draftsKey   = "drafts"
	errorKey    = "error"
	faultsKey   = "faults"
	warningsKey = "warnings"
	// whether IndexNow is told what changed
	indexNowToldKey = "indexnow"
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
		return check(arguments[contentArgument], os.Stderr)
	case command == checkCommand && len(arguments) == releaseArguments:
		folders := release.Folders{Content: arguments[contentArgument], Build: arguments[buildArgument], Cards: arguments[releaseCardsArgument]}
		return checkRelease(folders, os.Stderr)
	case command == serveCommand:
		asked, isWritten := servingOf(arguments[contentArgument:])
		if !isWritten {
			return usageWrong()
		}
		return serve(asked)
	case command == cardsCommand:
		asked, isWritten := cardsAskedOf(arguments[contentArgument:])
		if !isWritten {
			return usageWrong()
		}
		return drawCards(asked)
	case command == reviseCommand:
		asked, isWritten := postAskedOf(arguments[contentArgument:])
		if !isWritten {
			return usageWrong()
		}
		return revise(asked)
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
		return servingAsked{folders: release.Folders{Content: "", Build: "", Cards: ""}, drafts: false}, false
	}
	folders := release.Folders{Content: flags.Arg(contentPlace), Build: flags.Arg(buildPlace), Cards: flags.Arg(servedCardsPlace)}
	return servingAsked{folders: folders, drafts: *drafts}, true
}

// commandOf is the command the arguments name, or nothing where they name none.
func commandOf(arguments []string) string {
	if len(arguments) <= commandArgument {
		return ""
	}
	return arguments[commandArgument]
}

// check loads a content folder as the site would, and holds what it loads to the site's own check, as a release's
// load does, printing to warned what it warns of where it loads.
func check(contentFolder string, warned io.Writer) int {
	loaded, err := release.ReadFolder(contentFolder, content.Load)
	if err == nil {
		err = checkedNow(loaded)
	}
	if err == nil {
		warn(warned, loaded.Warnings, contentFolder)
	}
	return checked(err, contentFolder)
}

// checkRelease loads a release's folders as the site would, printing to warned what its content warns of where it
// loads.
func checkRelease(folders release.Folders, warned io.Writer) int {
	loaded, err := release.Load(folders, releaseChecks())
	if err == nil {
		warn(warned, loaded.Content.Warnings, folders.Content)
	}
	return checked(err, folders.Paths()...)
}

// warn prints what a content folder holds wrong that refuses nothing, after the folder, quoted, one a line; nothing
// where it holds none. The check succeeds all the same, so a commit and CI pass, and print it.
func warn(warned io.Writer, warnings content.Faults, contentFolder string) {
	if len(warnings) > 0 {
		// a warning that cannot be printed changes nothing the check decides
		_, _ = fmt.Fprintf(warned, warningsFormat, strconv.Quote(contentFolder), warnings)
	}
}

// checkedNow holds content to the site's own check, from the moment it is loaded on.
//
//wrapper:reads the clock as each load is checked, which the check takes as a value
func checkedNow(loaded *content.Snapshot) error {
	return compose.Check(loaded, time.Now()) //nolint:wrapcheck // the content's faults, printed one a line as they come
}

// postAsked is the post revise is asked of: the content folder, and the post's page, in the language asked.
type postAsked struct {
	folder string
	page   document.Page
}

// postAskedOf reads the post revise is asked of from the words after its command, and whether they are written right:
// the flag of its language, where one is asked, then the content folder and the post's slug. A mistake is told of by
// the usage alone, as for any other command.
func postAskedOf(words []string) (postAsked, bool) {
	flags := flag.NewFlagSet(reviseCommand, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	language := flags.String(languageFlag, "", languageUsage)
	if err := flags.Parse(words); err != nil || flags.NArg() != postArguments {
		return postAsked{folder: "", page: document.PageOf(document.PageNone)}, false
	}
	return postAsked{folder: flags.Arg(contentPlace), page: document.PostPage(flags.Arg(slugPlace)).InLanguage(*language)}, true
}

// postName is how a fault names a post: its slug, quoted, and the language of a translation after it.
func postName(post document.Page) string {
	if post.Translation == "" {
		return strconv.Quote(post.Slug)
	}
	return fmt.Sprintf(translationName, post.Slug, post.Translation)
}

// revise takes a revision of a post of a content folder out now, and prints the path it wrote it at.
func revise(asked postAsked) int {
	written, err := reviseAt(asked.folder, asked.page, time.Now())
	if err != nil {
		return refused(err, asked.folder)
	}
	fmt.Println(written)
	return succeeded
}

// reviseAt takes a revision of the current version of a post of a content folder out at a moment, beside the post,
// and dates the post as updated on the moment's day (docs/FOUNDATION.md §5.4, Revisions): the path it wrote, in the
// folder; a fault where the folder does not load, where the post is then no post out, a draft, one dated later or none
// at all, whose version no reader read, its version came out that day or later, a revision taken that day among them,
// or its text is its newest revision's already.
func reviseAt(contentFolder string, page document.Page, moment time.Time) (string, error) {
	loaded, err := release.ReadFolder(contentFolder, content.Load)
	if err != nil {
		return "", err //nolint:wrapcheck // the content's faults, printed one a line as they come
	}
	out := false
	for _, blog := range compose.SiteOf(compose.Prepare(loaded), compose.Audience{At: moment, Drafts: false}).Blogs() {
		out = out || slices.ContainsFunc(blog.Posts, func(post compose.Composed) bool { return post.Document.Head.Page == page })
	}
	post, found := loaded.Post(page.Translation, page.Slug)
	if !found || !out {
		return "", fmt.Errorf("%s is no post out now of the folder %q: a revision is taken of a version a reader read", postName(page), contentFolder)
	}
	taken, err := content.RevisionOf(post, moment)
	if err != nil {
		return "", fmt.Errorf("revise %s: %w", postName(page), err)
	}
	return taken.Path, saveRevision(contentFolder, post, taken, moment)
}

// saveRevision writes a revision's file beside its post in a content folder, opened as a root, so no path leads out of
// it, then dates the post as updated on the moment's day, so the folder loads again at once.
func saveRevision(contentFolder string, post content.Post, taken content.RevisionFile, moment time.Time) error {
	root, err := os.OpenRoot(contentFolder)
	if err != nil {
		return fmt.Errorf("open the content folder %q: %w", contentFolder, err)
	}
	err = writeRevisionIn(root, taken)
	if err == nil {
		err = datePostIn(root, post, moment)
	}
	return errors.Join(err, root.Close())
}

// writeRevisionIn writes a revision's file in a root, making its revisions folder where it is missing, never over a
// file there, and takes it away where it could not be written whole. A reload that meets the file half written refuses
// it as it refuses the whole one, which the post's current version's day names, until the post is dated by a later
// update.
func writeRevisionIn(root *os.Root, taken content.RevisionFile) error {
	if err := root.MkdirAll(path.Dir(taken.Path), writtenFolderMode); err != nil {
		return fmt.Errorf("make the revisions folder of %q: %w", taken.Path, err)
	}
	file, err := root.OpenFile(taken.Path, newFileOnly, writtenFileMode)
	if err != nil {
		return fmt.Errorf("make the revision %q: %w", taken.Path, err)
	}
	_, err = file.Write(taken.Text)
	if err = errors.Join(err, file.Close()); err != nil {
		return fmt.Errorf("write the revision %q: %w", taken.Path, errors.Join(err, root.Remove(taken.Path)))
	}
	return nil
}

// datePostIn dates a post's file in a root as updated on the day a moment falls on in the post's zone, every other byte
// as its author wrote it, and writes it in its place whole.
func datePostIn(root *os.Root, post content.Post, moment time.Time) error {
	file := content.PostFile(post)
	written, err := root.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read the post %q: %w", file, err)
	}
	dated, err := content.UpdatedText(post, written, moment)
	if err != nil {
		return fmt.Errorf("date the post %q: %w", file, err)
	}
	info, err := root.Stat(file)
	if err != nil {
		return fmt.Errorf("read the mode of the post %q: %w", file, err)
	}
	return replaceIn(root, file, dated, info.Mode().Perm())
}

// What a file written in place of another is called beside it until it takes the other's place: the other's name and
// this ending, which a reload refuses as a stray file of a post's folder where one is left, and the next drawing of the
// cards takes away from theirs
const replacementEnding = ".replacing"

// How a file written in place of another is opened beside it: to write, made new, or emptied where one was left
const replacementOnly = os.O_WRONLY | os.O_CREATE | os.O_TRUNC

// replaceIn writes a file of a root in its place whole: beside it first, kept on the disk, then renamed over it, so a
// reader of the folder, a reload among them, meets the file as it was or as written, never half written; the one beside
// it is taken away where it could not take the file's place.
func replaceIn(root *os.Root, file string, text []byte, mode os.FileMode) error {
	beside := file + replacementEnding
	written, err := root.OpenFile(beside, replacementOnly, mode)
	if err != nil {
		return fmt.Errorf("make %q: %w", beside, err)
	}
	_, err = written.Write(text)
	err = errors.Join(err, written.Sync(), written.Close())
	if err == nil {
		err = root.Rename(beside, file)
	}
	if err != nil {
		return fmt.Errorf("write %q in place of %q: %w", beside, file, errors.Join(err, root.Remove(beside)))
	}
	return nil
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

func serve(asked servingAsked) int {
	stopping, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	folders := asked.folders
	// The watch looks before the first load, so a change made while the site starts is a change it sees
	changes := release.Watch(stopping, folders.Paths(), lookInterval)
	store, err := release.Open(folders, releaseChecks())
	if err != nil {
		return refused(err, folders.Paths()...)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	indexNowKey, err := indexNowKeyOf(asked.drafts)
	if err != nil {
		logger.Error("the IndexNow key is refused", errorKey, err)
		return failed
	}
	settings := ssh.Settings{
		HostKeyPath: fromEnvironment(hostKeyVariable, defaultHostKey), OpeningTimeout: openingTimeout,
		IdleTimeout: sessionIdleTimeout, MaxTimeout: sessionMaxTimeout, Drafts: asked.drafts,
	}
	// both hosts compose from one prepared content, so each snapshot's code is split once
	prepared := release.DeriveFromContent(store, compose.Prepare)
	sshServer, err := ssh.New(store, prepared, logger, settings)
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
	webServer := newServer(web.New(store, prepared, asked.drafts, indexNowKey), logger)
	servers := []server{webServer, sshServer}
	// the notifier reads the store's current release as it wakes, so a wake still waiting stands for every reload after
	// it, and one place is all a reload needs
	reloaded := make(chan struct{}, 1)
	running := serving{
		logger:      logger,
		store:       store,
		servers:     servers,
		reloads:     make(chan os.Signal, 1),
		reloaded:    reloaded,
		changes:     changes,
		failures:    make(chan error, len(servers)),
		stopSignals: stop,
	}
	signal.Notify(running.reloads, syscall.SIGHUP)
	go func() { running.failures <- webServer.Serve(webListener) }()
	go func() { running.failures <- sshServer.Serve(sshListener) }()
	if indexNowKey != "" {
		notifier := indexnow.New(store, prepared, logger, indexnow.Settings{Key: indexNowKey, Endpoint: indexNowEndpoint, Client: &http.Client{}})
		go notifier.Run(stopping, reloaded)
	}
	logger.Info("serving", webKey, webAddress, sshKey, sshAddress, contentKey, folders.Content, buildKey, folders.Build, cardsKey, folders.Cards,
		draftsKey, asked.drafts, indexNowToldKey, indexNowKey != "")
	return running.watch(stopping)
}

// indexNowKeyOf is the IndexNow key the site is known by: the environment's, held to IndexNow's form; none where it
// sets none, or where the site serves its drafts to their author, which no search engine is told of.
func indexNowKeyOf(drafts bool) (string, error) {
	key := os.Getenv(indexNowKeyVariable)
	if drafts || key == "" {
		return "", nil
	}
	if err := indexnow.CheckKey(key); err != nil {
		return "", fmt.Errorf("%s: %w", indexNowKeyVariable, err)
	}
	return key, nil
}

// logWarnings logs what the content served holds wrong that refuses nothing, where it holds any.
func (running serving) logWarnings() {
	if warnings := running.store.Current().Content.Warnings; len(warnings) > 0 {
		running.logger.Warn("the content warns, refusing nothing", warningsKey, warnings.Error())
	}
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
	// reloaded tells IndexNow's notifier, where one runs, that a reload took
	reloaded chan<- struct{}
	// changes says the folders changed and held still, as the release's watch saw them
	changes  <-chan struct{}
	failures chan error
	// stopSignals gives SIGINT and SIGTERM back, so a second one ends the program while the first one waits
	stopSignals context.CancelFunc
}

// watch answers each signal until the site stops: a reload, a change in the folders, a stop, or the server failing;
// first it logs what the release it starts with warns of.
func (running serving) watch(stopping context.Context) int {
	running.logWarnings()
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
	release.SendUnlessWaiting(running.reloaded)
	running.logWarnings()
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
