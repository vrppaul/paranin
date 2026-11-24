package web

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/vrppaul/paranin/internal/document"
	"github.com/vrppaul/paranin/internal/routes"
)

// The fewest bytes a preview key holds: as many as SHA-256 writes, since a key shorter than its hash weakens the HMAC
// (RFC 2104, section 3)
const leastKeyBytes = sha256.Size

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

// Token is the token of a post's preview link, one segment of its path: the key's HMAC-SHA256 over the post's address
// once out, in base64url with no padding, so no address is changed without the key and a link opens one language
// alone, until the post is out or the key changes.
func (key *PreviewKey) Token(post document.Page) string {
	signer := hmac.New(sha256.New, key.secret)
	// a hash takes every write, and says so by returning no error
	_, _ = signer.Write([]byte(routes.Path(post, routes.FormatChosen)))
	return base64.RawURLEncoding.EncodeToString(signer.Sum(nil))
}

// admits is whether a token opens the preview link of a post: the key's token for the post; never where the server
// holds no key. It is compared in constant time, so how long a request takes tells nothing of how near its token came.
func (key *PreviewKey) admits(token string, post document.Page) bool {
	return key != nil && hmac.Equal([]byte(token), []byte(key.Token(post)))
}
