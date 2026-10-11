package editor

import (
	"fmt"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/layout/expr"
)

// SetBinding ties a property of the component to a named screen input, or
// unties it when name is empty. The property's current value stays as the
// input's default. It is one undo step; a locked component refuses.
func (e *Editor) SetBinding(id design.NodeID, key, name string) error {
	n, err := e.screenNode(id)
	if err != nil {
		return err
	}
	def, ok := e.cat.Get(n.Component)
	if !ok {
		return fmt.Errorf("editor: unknown component %q", n.Component)
	}
	if _, ok := def.Prop(key); !ok {
		return fmt.Errorf("editor: %s has no property %q", def.ID, key)
	}
	name, err = inputName(name)
	if err != nil {
		return err
	}
	if n.Bind[key] == name {
		return nil
	}
	e.apply(func() bool {
		return e.doc.Update(id, func(n *design.Node) {
			if name == "" {
				delete(n.Bind, key)
				if len(n.Bind) == 0 {
					n.Bind = nil
				}
				return
			}
			if n.Bind == nil {
				n.Bind = map[string]string{}
			}
			n.Bind[key] = name
		})
	})
	return nil
}

// SetShowIf makes the component appear only while the named yes/no input is
// true; an empty name makes it always appear. One undo step.
func (e *Editor) SetShowIf(id design.NodeID, name string) error {
	n, err := e.screenNode(id)
	if err != nil {
		return err
	}
	if expr.IsCondition(name) {
		// A condition on the window and on inputs: "w >= 100", "$Count > 0".
		cond, err := expr.ParseCond(name)
		if err != nil {
			return fmt.Errorf("editor: %w", err)
		}
		name = cond.String()
	} else if name, err = inputName(name); err != nil {
		return err
	}
	if n.ShowIf == name {
		return nil
	}
	e.apply(func() bool {
		return e.doc.Update(id, func(n *design.Node) { n.ShowIf = name })
	})
	return nil
}

// SetEvent names the event the screen raises when the component is clicked;
// an empty name raises none. One undo step.
func (e *Editor) SetEvent(id design.NodeID, name string) error {
	n, err := e.screenNode(id)
	if err != nil {
		return err
	}
	name, err = inputName(name)
	if err != nil {
		return err
	}
	if n.Event == name {
		return nil
	}
	e.apply(func() bool {
		return e.doc.Update(id, func(n *design.Node) { n.Event = name })
	})
	return nil
}

// SetKeys replaces the screen's key bindings with the ones written in spec
// ("s=Save:save, esc=Back"; see design.ParseKeys). One undo step.
func (e *Editor) SetKeys(spec string) error {
	keys, err := design.ParseKeys(spec)
	if err != nil {
		return fmt.Errorf("editor: %w", err)
	}
	if design.FormatKeys(keys) == design.FormatKeys(e.doc.Keys) {
		return nil
	}
	e.apply(func() bool {
		e.doc.Keys = keys
		return true
	})
	return nil
}

func (e *Editor) screenNode(id design.NodeID) (design.Node, error) {
	n, ok := e.doc.Get(id)
	if !ok {
		return n, fmt.Errorf("editor: no node %q", id)
	}
	if n.Locked {
		return n, fmt.Errorf("editor: %s is locked", n.Name)
	}
	return n, nil
}

// inputName cleans a name; empty stays empty (it clears).
func inputName(s string) (string, error) {
	s = design.CleanInputName(s)
	if s == "" {
		return "", nil
	}
	if !design.ValidInputName(s) {
		return "", fmt.Errorf("editor: a name starts with a letter and has letters, digits, spaces, - or _")
	}
	return s, nil
}
