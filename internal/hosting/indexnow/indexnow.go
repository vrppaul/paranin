// Package indexnow tells the search engines that read IndexNow, Bing, Yandex, Seznam and Naver among them, which pages
// to read again as a post comes out or changes, at once, where a sitemap waits for their next visit (docs/SEO.md,
// Crawling; indexnow.org). It is the one host that talks outward: it watches what a search engine may index of the site
// the public reads, and sends what changed since it last looked, as a reload takes or a post dated later comes out.
package indexnow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/vrppaul/paranin/internal/compose"
	"github.com/vrppaul/paranin/internal/release"
	"github.com/vrppaul/paranin/internal/routes"
)

// The form of a key IndexNow takes: 8 to 128 characters, each a letter or a digit of ASCII or a hyphen
const (
	shortestKey   = 8
	longestKey    = 128
	keyCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
)

// How a submission is sent: as JSON, by its header, and for how long its answer is waited for. IndexNow answers at
// once; ten seconds outlast a slow link and keep a dead endpoint from holding up the next change for long.
const (
	contentTypeHeader = "Content-Type"
	submissionType    = "application/json; charset=utf-8"
	answerTimeout     = 10 * time.Second
)

// The keys of what the log says of a submission that failed: the status IndexNow answered with, the error that kept
// it from answering, and how many addresses it held
const (
	statusKey    = "status"
	errorKey     = "error"
	addressesKey = "addresses"
)

// CheckKey refuses a key of another form than IndexNow's, saying why.
func CheckKey(key string) error {
	for _, character := range key {
		if !strings.ContainsRune(keyCharacters, character) {
			return fmt.Errorf("an IndexNow key holds letters and digits of ASCII and hyphens alone, and this one holds %q", character)
		}
	}
	// every character is of ASCII now, one byte each
	if length := len(key); length < shortestKey || length > longestKey {
		return fmt.Errorf("an IndexNow key is %d to %d characters long, and this one is %d", shortestKey, longestKey, length)
	}
	return nil
}

// Settings are what a notifier tells IndexNow with: the key the site is known by, whose file it serves at its root;
// where a submission goes; and the client it goes by.
type Settings struct {
	Key      string
	Endpoint string
	Client   *http.Client
}

// Notifier tells IndexNow, as the site changes, each address a search engine should read again.
type Notifier struct {
	store    *release.Store
	prepared *release.Derived[*compose.Prepared]
	logger   *slog.Logger
	settings Settings
}

// New is a notifier of the site the store holds, composed from each snapshot as prepared reads it, logging a submission
// that failed through the logger.
func New(store *release.Store, prepared *release.Derived[*compose.Prepared], logger *slog.Logger, settings Settings) *Notifier {
	return &Notifier{store: store, prepared: prepared, logger: logger, settings: settings}
}

// Run tells IndexNow what changed until the context ends. What is indexed as it starts is known already, so a start
// tells nothing; then, each time a reload takes, as reloads says, and each time a post dated later comes out, it sends
// every address added, whose day changed or gone since it last looked, in one submission, and none where none did.
// A submission that failed is logged and never sent again: the next change sends its own.
func (notifier *Notifier) Run(ctx context.Context, reloads <-chan struct{}) {
	known := notifier.indexedAt(time.Now())
	for {
		select {
		case <-ctx.Done():
			return
		case <-reloads:
		case <-changeAt(known.until):
		}
		latest := notifier.indexedAt(time.Now())
		if changed := latest.changedSince(known); len(changed) > 0 {
			notifier.submit(ctx, latest.addresses, changed)
		}
		known = latest
	}
}

// changeAt is when the site next changes by the clock, as a post dated later comes out: never, where none comes later.
func changeAt(until time.Time) <-chan time.Time {
	if until.IsZero() {
		return nil
	}
	return time.After(time.Until(until))
}

// indexed is what a search engine may index of the site the public reads at a moment: each page by its whole address
// and the day its text last changed, the zero time where it has none, as the sitemap names them; the site's addresses;
// and when a post dated later next changes it, the zero time where none comes later.
type indexed struct {
	days      map[string]time.Time
	addresses routes.Site
	until     time.Time
}

// indexedAt is what a search engine may index of the store's current snapshot at a moment, composed for the public.
func (notifier *Notifier) indexedAt(moment time.Time) indexed {
	snapshot := notifier.store.Current()
	site := compose.SiteOf(notifier.prepared.Of(snapshot, moment), compose.Audience{At: moment, Drafts: false})
	days := map[string]time.Time{}
	for _, page := range site.IndexedPages() {
		days[site.Addresses.Address(page.Page, routes.FormatChosen)] = page.Changed
	}
	return indexed{days: days, addresses: site.Addresses, until: site.Until}
}

// changedSince are the addresses that changed since what was known, in their order: each added, each whose day
// changed, and each gone from the site's host. One gone with a host the site left is not the site's to tell of, and
// IndexNow refuses a whole submission that names an address on a host other than the one it names.
func (latest indexed) changedSince(known indexed) []string {
	changed := []string{}
	for address, day := range latest.days {
		if knownDay, isKnown := known.days[address]; !isKnown || !knownDay.Equal(day) {
			changed = append(changed, address)
		}
	}
	host := latest.addresses.Host()
	for address := range known.days {
		if _, stays := latest.days[address]; !stays && isOnHost(address, host) {
			changed = append(changed, address)
		}
	}
	slices.Sort(changed)
	return changed
}

// isOnHost is whether an address leads to a host.
func isOnHost(address, host string) bool {
	parsed, err := url.Parse(address)
	return err == nil && parsed.Host == host
}

// submission is what IndexNow is sent (indexnow.org, "Submitting a set of URLs"): the site's host, the key it is known
// by and where the key's file is, and the addresses to read again.
type submission struct {
	Host        string   `json:"host"`
	Key         string   `json:"key"`
	KeyLocation string   `json:"keyLocation"`
	URLList     []string `json:"urlList"`
}

// submit tells IndexNow the addresses that changed of a site, and logs a submission that failed: one it was not sent,
// or one it answered with a status not of success.
func (notifier *Notifier) submit(ctx context.Context, addresses routes.Site, changed []string) {
	key := notifier.settings.Key
	status, err := notifier.send(ctx, submission{Host: addresses.Host(), Key: key, KeyLocation: addresses.KeyAddress(key), URLList: changed})
	switch {
	case err != nil:
		notifier.logger.Error("IndexNow could not be told what changed", errorKey, err, addressesKey, len(changed))
	case status < http.StatusOK || status >= http.StatusMultipleChoices:
		notifier.logger.Error("IndexNow refused what changed", statusKey, status, addressesKey, len(changed))
	}
}

// send sends a submission to the endpoint and is the status it is answered with, waiting for no longer than an answer
// takes; an error where it was not sent or not answered.
func (notifier *Notifier) send(ctx context.Context, sent submission) (int, error) {
	body, err := json.Marshal(sent)
	if err != nil {
		return 0, fmt.Errorf("write the submission: %w", err)
	}
	answering, cancel := context.WithTimeout(ctx, answerTimeout)
	defer cancel()
	endpoint := notifier.settings.Endpoint
	request, err := http.NewRequestWithContext(answering, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("make the submission to %q: %w", endpoint, err)
	}
	request.Header.Set(contentTypeHeader, submissionType)
	response, err := notifier.settings.Client.Do(request)
	if err != nil {
		return 0, fmt.Errorf("send the submission to %q: %w", endpoint, err)
	}
	// the answer's body says nothing a notifier acts on, and a close that fails loses nothing it reads
	_ = response.Body.Close()
	return response.StatusCode, nil
}
