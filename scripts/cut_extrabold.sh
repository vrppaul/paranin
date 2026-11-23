#!/bin/sh
# Cuts JetBrains Mono's ExtraBold, as JetBrains released it, to the characters of a mail address, the one text the page
# draws at 800 (design/tokens.yaml, mail), so a reader fetches only the glyphs a mail can use. It keeps printable ASCII,
# U+0020 to U+007E, all RFC 5321 lets an address hold, a quoted local part and an address literal among them; a letter
# past that range, as RFC 6531 allows, draws in the face's stand-in. Each glyph keeps its metrics and hints, and the
# font its names and licence; its contextual alternates, the code ligatures, go, since the page draws with ligatures
# off (web/src/ui/ui.css, body).
#
# Run in the repository's root: scripts/cut_extrabold.sh. It writes design/fonts/jetbrains-mono/
# JetBrainsMono-ExtraBold-ASCII.woff2 from the released JetBrainsMono-ExtraBold.woff2 beside it. It prints only what uvx
# prints on standard error when it fetches the pinned tools, and what pyftsubset prints when it fails. It exits 0 when
# the cut is written, 127 when uvx is missing, and otherwise with the code of uvx, where it cannot fetch the tools, or of
# pyftsubset, where it cannot cut.
set -eu

folder=design/fonts/jetbrains-mono

exec uvx --from fonttools==4.68.0 --with brotli==1.2.1 pyftsubset "$folder/JetBrainsMono-ExtraBold.woff2" \
	--unicodes=U+0020-007E \
	--layout-features-=calt \
	--name-IDs='*' \
	--flavor=woff2 \
	--output-file="$folder/JetBrainsMono-ExtraBold-ASCII.woff2"
