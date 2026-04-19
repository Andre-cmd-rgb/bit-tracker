package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type PetState int

const (
	PetHappy PetState = iota
	PetExcited
	PetSad
	PetSleepy
	PetFocused
)

// Wall-E inspired: boxy head with two big round eye-lenses on a stalk,
// trapezoid body and chunky track treads. Four frames per state for smooth
// animation (idle sway + blink + tread shuffle). Each frame is 16 wide x 10 tall.
var petArt = map[PetState][4]string{
	PetHappy: {
		// frame 0 — eyes open, head slight left
		`   ╭──────╮
   │ ◉  ◉ │
   ╰──╥───╯
   ╔══╩═════╗
   ║ ·BIT·  ║
   ║  ┌──┐  ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃●●●●●●●●┃
   ┗━━━━━━━━┛  `,
		// frame 1 — head center, eyes slight look-right
		`    ╭──────╮
    │ ·◉ ◉·│
    ╰──╥───╯
    ╔══╩═════╗
    ║ ·BIT·  ║
    ║  ┌──┐  ║
    ╚══╧══╧══╝
    ┏━━━━━━━━┓
    ┃●●●●●●●●┃
    ┗━━━━━━━━┛ `,
		// frame 2 — blink
		`    ╭──────╮
    │ ─  ─ │
    ╰──╥───╯
    ╔══╩═════╗
    ║ ·BIT·  ║
    ║  ┌──┐  ║
    ╚══╧══╧══╝
    ┏━━━━━━━━┓
    ┃●●●●●●●●┃
    ┗━━━━━━━━┛ `,
		// frame 3 — wiggle right
		`     ╭──────╮
     │ ◉  ◉ │
     ╰──╥───╯
     ╔══╩═════╗
     ║ ·BIT·  ║
     ║  ┌──┐  ║
     ╚══╧══╧══╝
     ┏━━━━━━━━┓
     ┃●●●●●●●●┃
     ┗━━━━━━━━┛`,
	},
	PetExcited: {
		`  ✦ ╭──────╮  ✦
    │ ◕  ◕ │
    ╰──╥───╯
    ╔══╩═════╗
   ⚡ ·YAY·  ⚡
    ║  ▽▽▽  ║
    ╚══╧══╧══╝
    ┏━━━━━━━━┓
    ┃◆◆◆◆◆◆◆◆┃
    ┗━━━━━━━━┛ `,
		`  ✧ ╭──────╮ ✧
    │ ★  ★ │
    ╰──╥───╯
   ╔═══╩═════╗
   ║ ·WOW!·  ║
   ║   ▽▽    ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃◆◆◆◆◆◆◆◆┃
   ┗━━━━━━━━┛  `,
		` ✦  ╭──────╮  ✦
    │ ◉  ◉ │
    ╰──╥───╯
    ╔══╩═════╗
   ⚡·NICE!·⚡
    ║  ▽▽▽  ║
    ╚══╧══╧══╝
    ┏━━━━━━━━┓
    ┃◆◆◆◆◆◆◆◆┃
    ┗━━━━━━━━┛ `,
		`    ╭──────╮ ✧
  ✧ │ ◕  ◕ │
    ╰──╥───╯
   ╔═══╩═════╗
   ║ ·YAY·   ║
   ║  ▽▽▽▽   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃◆◆◆◆◆◆◆◆┃
   ┗━━━━━━━━┛  `,
	},
	PetSad: {
		`   ╭──────╮
   │ ╥  ╥ │
   ╰──╥───╯
   ╔══╩═════╗
   ║ ·hmm·  ║
   ║  ︵︵   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃░░░░░░░░┃
   ┗━━━━━━━━┛  `,
		`   ╭──────╮
   │ ╥  ╥ │   ·
   ╰──╥───╯
   ╔══╩═════╗
   ║ ·...·  ║
   ║  ︵︵   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃░░░░░░░░┃
   ┗━━━━━━━━┛  `,
		`   ╭──────╮
   │ —  — │
   ╰──╥───╯
   ╔══╩═════╗
   ║        ║
   ║  ︵︵   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃░░░░░░░░┃
   ┗━━━━━━━━┛  `,
		`  ·╭──────╮
   │ ╥  ╥ │
   ╰──╥───╯
   ╔══╩═════╗
   ║ ·miss· ║
   ║  ︵︵   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃░░░░░░░░┃
   ┗━━━━━━━━┛  `,
	},
	PetSleepy: {
		`   ╭──────╮  z
   │ —  — │   Z
   ╰──╥───╯
   ╔══╩═════╗
   ║ ·zzz·  ║
   ║  ~~~   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃····••··┃
   ┗━━━━━━━━┛  `,
		`   ╭──────╮
   │ —  — │  z
   ╰──╥───╯   Z
   ╔══╩═════╗
   ║ ·zz··  ║
   ║  ~~~   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃··••····┃
   ┗━━━━━━━━┛  `,
		`   ╭──────╮  Z
   │ —  — │
   ╰──╥───╯  z
   ╔══╩═════╗
   ║ ·z···  ║
   ║  ~~~   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃·••·····┃
   ┗━━━━━━━━┛  `,
		`   ╭──────╮
   │ ◜  ◜ │  z
   ╰──╥───╯   Z
   ╔══╩═════╗
   ║ ·yawn· ║
   ║  ~~~   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃·····••·┃
   ┗━━━━━━━━┛  `,
	},
	PetFocused: {
		`   ╭──────╮
   │ ◎  ◎ │
   ╰──╥───╯
   ╔══╩═════╗
   ║·FOCUS ·║
   ║  ━━━   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃▓▓▓▓▓▓▓▓┃
   ┗━━━━━━━━┛  `,
		`   ╭──────╮
   │ ◎  ◎ │ ⟶
   ╰──╥───╯
   ╔══╩═════╗
   ║·WRITE· ║
   ║  ━━━   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃▓▓▓▓▓▓▓▓┃
   ┗━━━━━━━━┛  `,
		`   ╭──────╮
   │ ◉  ◉ │
   ╰──╥───╯
   ╔══╩═════╗
   ║·THINK· ║
   ║  ━━━   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃▓▓▓▓▓▓▓▓┃
   ┗━━━━━━━━┛  `,
		`   ╭──────╮
   │ ◎  ◎ │⟵
   ╰──╥───╯
   ╔══╩═════╗
   ║·NOTE·· ║
   ║  ━━━   ║
   ╚══╧══╧══╝
   ┏━━━━━━━━┓
   ┃▓▓▓▓▓▓▓▓┃
   ┗━━━━━━━━┛  `,
	},
}

func petColor(s PetState) lipgloss.Color {
	switch s {
	case PetExcited:
		return lipgloss.Color("#FFD700")
	case PetSad:
		return lipgloss.Color("#778899")
	case PetSleepy:
		return lipgloss.Color("#9B8EC4")
	case PetFocused:
		return lipgloss.Color("#FF9F43")
	default:
		return lipgloss.Color("#7DF9AA")
	}
}

// accent picks a complementary highlight color for the eye-lenses / sparkles.
func accentColor(s PetState) lipgloss.Color {
	switch s {
	case PetExcited:
		return lipgloss.Color("#FFF1A0")
	case PetSad:
		return lipgloss.Color("#BDC3C7")
	case PetSleepy:
		return lipgloss.Color("#D6CCF0")
	case PetFocused:
		return lipgloss.Color("#FFD8A8")
	default:
		return lipgloss.Color("#D8FFE9")
	}
}

// RenderPet returns a styled, multi-color pet block. frame cycles 0..3.
func RenderPet(state PetState, frame int) string {
	art := petArt[state][frame%4]
	body := lipgloss.NewStyle().Foreground(petColor(state))
	accent := lipgloss.NewStyle().Foreground(accentColor(state)).Bold(true)

	lines := strings.Split(art, "\n")
	out := make([]string, len(lines))
	for i, l := range lines {
		rendered := make([]rune, 0, len(l))
		var buf strings.Builder
		// split into body + accent chars
		for _, r := range l {
			switch r {
			case '◉', '◕', '★', '◎', '✦', '✧', '⚡', '◆', '◜':
				if buf.Len() > 0 {
					out[i] += body.Render(buf.String())
					buf.Reset()
				}
				out[i] += accent.Render(string(r))
			default:
				buf.WriteRune(r)
			}
			_ = rendered
		}
		if buf.Len() > 0 {
			out[i] += body.Render(buf.String())
		}
	}
	return strings.Join(out, "\n")
}

// PetLabel returns a short mood label.
func PetLabel(s PetState) string {
	switch s {
	case PetExcited:
		return "excited"
	case PetSad:
		return "missing you"
	case PetSleepy:
		return "sleepy"
	case PetFocused:
		return "focused"
	default:
		return "happy"
	}
}
