package tui

import "strings"

// The room in a row of the menu over the command line: a cell either side of an item, as Neovim's popup menu leaves
// it, and four between its name and its kind, as the owner's storyboard parts them
const (
	menuPadding = 1
	kindGap     = 4
)

// menuWidths are the widths of the widest name and the widest kind among what Tab goes round.
type menuWidths struct {
	name int
	kind int
}

// menuColumns are the columns every row of the menu takes on the screen as it is, so its names and its kinds stand
// in columns: the widest name's and kind's; whether the screen has room for the kinds; and whether the menu holds too
// many items to show at once, a scrollbar ending each row.
type menuColumns struct {
	widest   menuWidths
	kinds    bool
	scrolled bool
}

// itemParts are the parts an item of the menu is painted in: its name, the letters of it the reader typed, its kind.
type itemParts struct {
	name  part
	typed part
	kind  part
}

// menuWidthsOf are the widths of the widest name and kind among what Tab goes round.
func menuWidthsOf(listed []completion) menuWidths {
	widest := menuWidths{name: 0, kind: 0}
	for _, item := range listed {
		widest = menuWidths{name: max(widest.name, textWidth(item.name)), kind: max(widest.kind, textWidth(item.kind))}
	}
	return widest
}

// firstInSight is the first item a menu with room for so many shows of so many: the one it showed, moved only as far
// as keeps the item chosen in sight, and the menu full.
func firstInSight(first, chosen, room, count int) int {
	if chosen != noneChosen {
		first = min(max(first, chosen-room+1), chosen)
	}
	return max(min(first, count-room), 0)
}

// withMenu lays the menu of what Tab goes round over the page's rows, where anything fits what is typed: its last row
// over the command line and its names under where the command line writes them, held within the screen; as many items
// as the page has rows, the one chosen among them, and a scrollbar where some are left out.
func (model Model) withMenu(rows []string, command commandLine) []string {
	listed, room := model.completion.listed, model.frame.rows
	line, shown := model.frame.commandRow()
	if len(listed) == 0 || room == 0 || !shown {
		return rows
	}
	columns := model.menuColumns()
	first, sight := model.completion.first, min(room, len(listed))
	thumbRows := thumbSpan{start: 0, end: 0}
	if columns.scrolled {
		thumbRows = thumbOf(room, len(listed), first)
	}
	painted, width := make([]string, 0, sight), 0
	for row := range sight {
		place := first + row
		pieces := model.menuItem(listed[place], place == model.completion.chosen, columns)
		if columns.scrolled {
			pieces = append(pieces, scrollMark(thumbRows.covers(row)))
		}
		// every row is as wide, its columns padded to their widest
		bar := model.paint.cut(pieces, model.frame.width)
		painted, width = append(painted, bar.text), bar.width
	}
	left := command.typedEnd - textWidth(model.writtenName()) - menuPadding
	return laidOver(rows, painted, line-sight, min(max(left, 0), model.frame.width-width), model.frame.width)
}

// writtenName is what the command line writes last of what Tab goes round: the name Tab chose, or the letters of a
// name the reader typed.
func (model Model) writtenName() string {
	if model.completion.chosen == noneChosen {
		return model.completion.matched
	}
	return model.completion.listed[model.completion.chosen].name
}

// scrollMark is the cell of the menu's scrollbar on a row of it: the thumb, where it takes the row, or the track.
func scrollMark(onThumb bool) piece {
	if onThumb {
		return piece{text: thumbMark, part: partMenuThumb}
	}
	return piece{text: trackMark, part: partMenuTrack}
}

// menuColumns are the columns of the menu over the command line on the screen as it is: the kinds left out where
// they would take the menu past the screen's width.
func (model Model) menuColumns() menuColumns {
	widest, scrolled := model.completion.widest, len(model.completion.listed) > model.frame.rows
	ending := menuPadding
	if scrolled {
		ending = textWidth(space + trackMark)
	}
	kinds := menuPadding+widest.name+kindGap+widest.kind+ending <= model.frame.width
	return menuColumns{widest: widest, kinds: kinds, scrolled: scrolled}
}

// menuItem is the pieces of a row of the menu, but for its scrollbar: the item's name, the letters of it typed
// marked, and its kind, each padded to its column; all of it marked as chosen where Tab chose the item.
func (model Model) menuItem(item completion, chosen bool, columns menuColumns) []piece {
	parts := itemParts{name: partMenu, typed: partMenuTyped, kind: partMenuKind}
	if chosen {
		parts = itemParts{name: partMenuChosen, typed: partMenuChosen, kind: partMenuChosen}
	}
	typed := model.completion.matched
	pieces := []piece{
		{text: strings.Repeat(space, menuPadding), part: parts.name},
		{text: typed, part: parts.typed},
		{text: padded(strings.TrimPrefix(item.name, typed), columns.widest.name-textWidth(typed)), part: parts.name},
	}
	if columns.kinds {
		pieces = append(pieces, piece{text: strings.Repeat(space, kindGap) + padded(item.kind, columns.widest.kind), part: parts.kind})
	}
	if columns.scrolled {
		return append(pieces, piece{text: space, part: parts.name})
	}
	return append(pieces, piece{text: strings.Repeat(space, menuPadding), part: parts.name})
}
