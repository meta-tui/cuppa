package editor

import (
	"fmt"
	"reflect"

	"github.com/meta-ui-labs/cuppa/libs/color/space"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Effect names accepted by SetEffect.
const (
	EffectGrid      = "grid"
	EffectShadow    = "shadow"
	EffectScanlines = "scanlines"
	EffectVignette  = "vignette"
)

// SetCanvasSize resizes the canvas, as one undo step. Components with layout
// expressions follow the new size; the others stay where they are, and are just
// not drawn if outside it until it grows again.
func (e *Editor) SetCanvasSize(w, h int) error {
	if w < 1 || h < 1 || w > design.MaxWidth || h > design.MaxHeight {
		return fmt.Errorf("editor: the canvas is 1 to %d wide and 1 to %d high", design.MaxWidth, design.MaxHeight)
	}
	if w == e.doc.Width && h == e.doc.Height {
		return nil
	}
	e.apply(func() bool {
		e.doc.Width, e.doc.Height = w, h
		e.Resolve()
		return true
	})
	return nil
}

// SetBackground sets the canvas colour ("" for the terminal's own), as one undo
// step. The value is validated like a component colour.
func (e *Editor) SetBackground(value string) error {
	value, err := space.Normalise(value)
	if err != nil {
		return fmt.Errorf("editor: %w", err)
	}
	if value == e.doc.Background {
		return nil
	}
	e.apply(func() bool {
		e.doc.Background = value
		return true
	})
	return nil
}

// Theme roles accepted by SetThemeColor (the background is SetBackground).
const (
	ThemeText      = "text"
	ThemeMuted     = "muted"
	ThemeBorder    = "border"
	ThemeSecondary = "secondary"
)

// SetThemeColor sets one colour of the design's theme ("" for none: the
// component defaults show again), as one undo step. Components that set their
// own colour keep it.
func (e *Editor) SetThemeColor(role, value string) error {
	value, err := space.Normalise(value)
	if err != nil {
		return fmt.Errorf("editor: %w", err)
	}
	var field *string
	switch role {
	case ThemeText:
		field = &e.doc.Theme.Text
	case ThemeMuted:
		field = &e.doc.Theme.Muted
	case ThemeBorder:
		field = &e.doc.Theme.Border
	case ThemeSecondary:
		field = &e.doc.Theme.Secondary
	default:
		return fmt.Errorf("editor: no theme colour %q", role)
	}
	if *field == value {
		return nil
	}
	e.apply(func() bool {
		switch role {
		case ThemeText:
			e.doc.Theme.Text = value
		case ThemeMuted:
			e.doc.Theme.Muted = value
		case ThemeBorder:
			e.doc.Theme.Border = value
		case ThemeSecondary:
			e.doc.Theme.Secondary = value
		}
		return true
	})
	return nil
}

// SetTheme sets the background and the whole theme at once, as one undo step:
// a preset. Each colour is validated like any other. Components that set their
// own colour keep it.
func (e *Editor) SetTheme(background string, t design.Theme) error {
	var err error
	if background, err = space.Normalise(background); err != nil {
		return fmt.Errorf("editor: %w", err)
	}
	for _, c := range []*string{&t.Text, &t.Muted, &t.Border, &t.Secondary} {
		if *c, err = space.Normalise(*c); err != nil {
			return fmt.Errorf("editor: %w", err)
		}
	}
	t.Palette = e.doc.Theme.Palette // a colour scheme does not touch the named colours
	if background == e.doc.Background && reflect.DeepEqual(t, e.doc.Theme) {
		return nil
	}
	e.apply(func() bool {
		e.doc.Background, e.doc.Theme = background, t
		return true
	})
	return nil
}

// SetEffect turns one of the canvas options (grid, shadow, scanlines,
// vignette) on or off, as one undo step.
func (e *Editor) SetEffect(name string, on bool) error {
	if !knownEffect(name) {
		return fmt.Errorf("editor: no effect %q", name)
	}
	if e.Effect(name) == on {
		return nil
	}
	e.apply(func() bool {
		switch name {
		case EffectGrid:
			e.doc.HideGrid = !on
		case EffectShadow:
			e.doc.Effects.Shadow = on
		case EffectScanlines:
			e.doc.Effects.Scanlines = on
		case EffectVignette:
			e.doc.Effects.Vignette = on
		}
		return true
	})
	return nil
}

// SetProfile chooses the colour profile the design targets ("" for true
// colour, "256", "16" or "none"), as one undo step.
func (e *Editor) SetProfile(profile string) error {
	if !design.ValidProfile(profile) {
		return fmt.Errorf("editor: the colour profile is true colour, 256, 16 or none")
	}
	if profile == e.doc.Profile {
		return nil
	}
	e.apply(func() bool {
		e.doc.Profile = profile
		return true
	})
	return nil
}

// SetLight previews the design on a light (true) or dark (false) terminal, as
// one undo step.
func (e *Editor) SetLight(light bool) {
	if light == e.doc.Light {
		return
	}
	e.apply(func() bool {
		e.doc.Light = light
		return true
	})
}

// Effect reports whether a canvas option is on. The grid is on unless hidden.
func (e *Editor) Effect(name string) bool {
	switch name {
	case EffectGrid:
		return !e.doc.HideGrid
	case EffectShadow:
		return e.doc.Effects.Shadow
	case EffectScanlines:
		return e.doc.Effects.Scanlines
	case EffectVignette:
		return e.doc.Effects.Vignette
	}
	return false
}

func knownEffect(name string) bool {
	switch name {
	case EffectGrid, EffectShadow, EffectScanlines, EffectVignette:
		return true
	}
	return false
}
