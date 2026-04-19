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

// Each pet art is two frames of exactly 14 wide x 9 tall (not counting trailing newline).
var petArt = map[PetState][2]string{
	PetHappy: {
		`   ╭──────╮
  │  ^  ^ │
  │   ◡   │
  │ ╭──╮  │
  ╰─┴──┴──╯
   ││    ││
   ╰╯    ╰╯
     bit
              `,
		`   ╭──────╮
  │  ^  ^ │
  │   ◡   │
  │ ╰──╯  │
  ╰─┴──┴──╯
   ││    ││
   ╯╰    ╯╰
     bit
              `,
	},
	PetExcited: {
		`  ✦ ╭──────╮✦
  │  ◉  ◉ │
  │   ▽   │
  │  wow  │
  ╰─┴──┴──╯
   ││    ││
   ╰╯    ╰╯
   * party *
              `,
		` ✦  ╭──────╮ ✦
  │  ◉  ◉ │
  │   ▽   │
  │  yay! │
  ╰─┴──┴──╯
   ││    ││
   ╯╰    ╯╰
   * party *
              `,
	},
	PetSad: {
		`   ╭──────╮
  │  T  T │
  │   ︵   │
  │  hmm  │
  ╰─┴──┴──╯
   ││    ││
   ╰╯    ╰╯
    quiet
              `,
		`   ╭──────╮
  │  T  T │
  │   ︵   │
  │       │
  ╰─┴──┴──╯
   ││    ││
   ╰╯    ╰╯
    quiet
              `,
	},
	PetSleepy: {
		`   ╭──────╮  z
  │  -  - │
  │   ～   │
  │  zzz  │
  ╰─┴──┴──╯
   ││    ││
   ╰╯    ╰╯
    sleepy
              `,
		`  z╭──────╮
  │  -  - │  Z
  │   ～   │
  │  zzz  │
  ╰─┴──┴──╯
   ││    ││
   ╰╯    ╰╯
    sleepy
              `,
	},
	PetFocused: {
		`   ╭──────╮
  │  >  < │
  │   ―   │
  │ focus │
  ╰─┴──┴──╯
   ││    ││
   ╰╯    ╰╯
    writing
              `,
		`   ╭──────╮
  │  >  < │
  │   ―   │
  │ focus │
  ╰─┴──┴──╯
   ││    ││
   ╯╰    ╯╰
    writing
              `,
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

// RenderPet returns a styled pet block.
func RenderPet(state PetState, frame int) string {
	art := petArt[state][frame%2]
	st := lipgloss.NewStyle().Foreground(petColor(state))
	lines := strings.Split(art, "\n")
	for i, l := range lines {
		lines[i] = st.Render(l)
	}
	return strings.Join(lines, "\n")
}

// PetLabel returns a human name for the state.
func PetLabel(s PetState) string {
	switch s {
	case PetExcited:
		return "excited"
	case PetSad:
		return "sad"
	case PetSleepy:
		return "sleepy"
	case PetFocused:
		return "focused"
	default:
		return "happy"
	}
}
