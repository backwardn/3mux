package wm

import (
	"fmt"
	"strings"

	"github.com/aaronjanse/3mux/render"
)

// A Universe contains workspaces
type Universe struct {
	workspaces   []*workspace
	selectionIdx int
	renderRect   Rect
	renderer     *render.Renderer

	onDeath func(error)
}

func NewUniverse(renderer *render.Renderer, onDeath func(error), renderRect Rect, newPane NewPaneFunc) *Universe {
	u := &Universe{
		selectionIdx: 0,
		renderRect:   renderRect,
		onDeath:      onDeath,
		renderer:     renderer,
	}
	u.workspaces = []*workspace{newWorkspace(renderer, u.redrawAllLines, u.handleChildDeath, renderRect, newPane)}
	u.updateSelection()
	u.refreshRenderRect()
	return u
}

func (u *Universe) Serialize() string {
	out := fmt.Sprintf("Universe[%d]", u.selectionIdx)

	out += "("
	for i, e := range u.workspaces {
		if i != 0 {
			out += ", "
		}
		out += e.serialize()
	}
	out += ")"

	return out
}

// setRenderRect updates the Split's renderRect cache after which it calls refreshRenderRect
// this for when a split is reshaped
func (u *Universe) SetRenderRect(x, y, w, h int) {
	u.renderRect = Rect{x, y, w, h}

	// NOTE: should we clear the screen?

	u.refreshRenderRect()
}

func (u *Universe) getRenderRect() Rect {
	return u.renderRect
}

func (u *Universe) setPaused(pause bool) {
	for _, n := range u.workspaces {
		n.contents.SetPaused(pause)
	}
}

func (u *Universe) redrawAllLines() {
	for _, n := range u.workspaces {
		n.contents.redrawLines()
	}
}

// refreshRenderRect recalculates the coordinates of a Split's elements and calls setRenderRect on each of its children
// this is for when one or more of a split's children are reshaped
func (u *Universe) refreshRenderRect() {
	x := u.renderRect.X
	y := u.renderRect.Y
	w := u.renderRect.W
	h := u.renderRect.H

	for _, child := range u.workspaces {
		child.setRenderRect(x, y, w, h-2)
	}

	u.drawHelpBar()

	u.redrawAllLines()
	u.drawSelectionBorder()
}

func (u *Universe) drawHelpBar() {
	for _, hb := range helpBar {
		if helpBarMinLen(hb[0]) > u.renderRect.W {
			continue
		}

		widthStr := hb[0]
		widthStr = strings.ReplaceAll(widthStr, "{", "")
		widthStr = strings.ReplaceAll(widthStr, "}", "")

		space := " "
		for {
			test := strings.ReplaceAll(widthStr, "\t", space+" ")
			if len(test) >= u.renderRect.W {
				break
			}
			space += " "
		}

		style := render.Style{}

		for line := 0; line < 2; line++ {
			x := 0
			for _, r := range strings.ReplaceAll(hb[line], "\t", space) {
				// log.Printf("%q", r)
				switch r {
				case '{':
					style.Reverse = true
				case '}':
					style.Reverse = false
				default:
					u.renderer.HandleCh(render.PositionedChar{
						Rune: r,
						Cursor: render.Cursor{
							X: x, Y: u.renderRect.H - 2 + line,
							Style: style,
						},
					})
					x++
				}
			}
		}

		break
	}
}

func helpBarMinLen(str string) int {
	var x string
	x = strings.ReplaceAll(str, "{", "")
	x = strings.ReplaceAll(str, "}", "")
	x = strings.ReplaceAll(str, "\t", " ")
	return len(x)
}

var helpBar [][2]string = [][2]string{
	[2]string{
		"Alt+...      \t{N} New Pane  \t{Arrow} Move Pane     \t{/} Search    \t{\\} Hide Help",
		"Alt+Shift+...\t{Q} Close Pane\t{Arrow} Move Selection\t{F} Fullscreen\t{?} More Help",
	},
	[2]string{
		"Alt+...      \t{N} New Pane  \t{Arrow} Move Pane     \t{\\} Hide Help",
		"Alt+Shift+...\t{Q} Close Pane\t{Arrow} Move Selection\t{?} More Help",
	},
	[2]string{
		"Alt+...      \t{N} New Pane  \t{Arrow} Move  \t{\\} Hide Help",
		"Alt+Shift+...\t{Q} Close Pane\t{Arrow} Select\t{?} More Help",
	},
	[2]string{
		"Alt+...      \t{N} New  \t{Arrow} Move  \t{\\} Hide Help",
		"Alt+Shift+...\t{Q} Close\t{Arrow} Select\t{?} More Help",
	},
	[2]string{
		"Alt+...      \t{N} New  \t{\\} Hide Help",
		"Alt+Shift+...\t{Q} Close\t{?} More Help",
	},
	[2]string{
		"{Alt+X}  Hide Help",
		"{Alt+?}  More Help",
	},
	[2]string{
		"",
		"",
	},
}