package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The fewest bytes a preview key holds: as many as SHA-256 writes, since a key shorter than its hash weakens the HMAC
// (RFC 2104, section 3)
const leastKeyBytes = sha256.Size

// How a preview link's token is written, one segment of its path: the moment the link lapses, in seconds of Unix time,
// then its signature, the HMAC-SHA256 over that moment and the address the post has once out, in base64url with no
// padding; and what joins the moment to the address in what is signed. Neither mark stands in a number or base64url
const (
	tokenJoint  = "."
	signedJoint = "/"
	decimal     = 10
	secondsBits = 64
)

// PreviewKey is the secret a post's preview link is signed with (docs/FOUNDATION.md §5.4, Status), which the server and
// the owner's machine each hold and git never does; a new key ends every link signed with the one before.
type PreviewKey struct {
	secret []byte
}

// ReadPreviewKey reads a preview key from its file: what the file holds, the spaces and line ends round it left out, so
// a copy of it written with a line end or without signs alike; refused where it holds fewer bytes than leastKeyBytes.
func ReadPreviewKey(path string) (*PreviewKey, error) {
	//nolint:gosec // the path is the operator's own setting, and naming any file is what it is for
	written, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read the preview key %q: %w", path, err)
	}
	secret := bytes.TrimSpace(written)
	if len(secret) < leastKeyBytes {
		return nil, fmt.Errorf("the preview key %q holds %d bytes, fewer than %d", path, len(secret), leastKeyBytes)
	}
	return &PreviewKey{secret: secret}, nil
}

// Token is the token of a post's preview link that lapses at a moment: the moment, then the key's signature over it
// and the post's address once out, so neither is changed without the key and a link opens one language alone.
func (key *PreviewKey) Token(post document.Page, lapses time.Time) string {
	moment := strconv.FormatInt(lapses.Unix(), decimal)
	return moment + tokenJoint + key.signature(moment, routes.Path(post, routes.FormatChosen))
}

// admits is whether a token opens the preview link of a post at a moment: written as Token writes one, signed by the
// key over the post's address and the moment it lapses, which has not come; never where the server holds no key. The
// signature is compared in constant time, so how long a request takes tells nothing of how near its token came.
func (key *PreviewKey) admits(token string, post document.Page, moment time.Time) bool {
	if key == nil {
		return false
	}
	lapsesWritten, signature, isWritten := strings.Cut(token, tokenJoint)
	lapses, err := strconv.ParseInt(lapsesWritten, decimal, secondsBits)
	if !isWritten || err != nil || !moment.Before(time.Unix(lapses, 0)) {
		return false
	}
	return hmac.Equal([]byte(signature), []byte(key.signature(lapsesWritten, routes.Path(post, routes.FormatChosen))))
}

// signature is the key's HMAC-SHA256 over the moment a link lapses, as its token writes it, and a post's address.
func (key *PreviewKey) signature(lapses, address string) string {
	signer := hmac.New(sha256.New, key.secret)
	// a hash takes every write, and says so by returning no error
	_, _ = signer.Write([]byte(lapses + signedJoint + address))
	return base64.RawURLEncoding.EncodeToString(signer.Sum(nil))
}
