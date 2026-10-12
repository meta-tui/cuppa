# 9. Ports: changing a component for a compatible one

Status: superseded by ADR 0017 (removed in #271)

## Context

A component is bound to a screen's variables (ADR 0007): its properties to inputs, its visibility to a yes/no input, its click to an event. Changing a text input for a colour picker, or tabs for a row of buttons, meant deleting one and adding the other, and every variable was lost.

Two components can stand in for each other exactly when they carry the same kind of data. That has to be declared, not guessed from property names (`value`, `text` and `content` are all a string).

## Decision

### Ports

A property of a component may have a **port**: what it carries.

| Port | Carries | Property format |
|---|---|---|
| `value` | a string the person edits or the program shows (a colour is one) | text or colour |
| `items` | a list of items to choose from or to show | items separated by commas |
| `selected` | the index of the chosen item, the active tab or button | a number |
| `checked` | a yes/no answer | true or false |
| `headers` | the names of the columns of a table | names separated by commas |
| `rows` | the rows of a table | cells by commas, rows by semicolons |
| `outline` | lines with a depth, such as a tree | lines separated by `\|`, two spaces per level |
| `numbers` | a series for a chart | numbers separated by commas |
| `labels` | a label for each number | names separated by commas |
| `percent` | a number from 0 to 100 | a number |

The table is `libs/catalog/standard/ports_content.go`. A property with no port is only that component's (a border style, a font).

### Compatibility

`libs/catalog/compat` answers one question: when **from** is changed for **to**, which bound properties of **from** have a property of **to** with the same port? Those **move** (their binding, and their value when it is valid for the new property); the others are **lost**.

- A change is *compatible* when nothing bound is lost. A text input bound to `Colour` and a colour picker are compatible (both `value`); tabs bound to `Sections` and `Section` and dialog buttons are compatible (`items`, `selected`); a list bound to `Teas` and a label are not (a label has no `items`).
- A component with nothing bound can be changed for one that shares at least one port with it.
- Show-if, event, name, place, layout and the other properties never block a change: they belong to the node, not to the component.
- Only plain components change: a group, a pack component and a locked component do not.

### The change

`Editor.ChangeComponent(id, component, allowLoss)` is one undo step:

- name, position, size (made larger when the new component's smallest size needs it), layout, show-if, event, visibility and locking stay;
- the properties of a moved port keep their binding under the new property's key, and their value when it is valid there (a text input's `hello` is not a colour, so a picker starts from its own default and keeps the binding);
- other properties with the same key and kind keep their values;
- a change that would lose a binding is refused with the names of the variables, unless `allowLoss` is true.

`Editor.ChangeOptions(id)` lists the compatible components in catalog order; `ChangeReport` says what a change would keep, lose and add without doing it.

### Where it is offered

- The details bar has a *Change* row under *Type*: the arrows choose among the compatible components, pressing the name changes the component.
- `cuppa component options <design.cuppa> <node>` and `cuppa component change <design.cuppa> <node> <component> [--allow-loss] [-o out]` do the same from a script or an agent.

## Limits

- The typed screen input follows the new component: a text input's input is a `string`, a colour picker's is a `ColourPicker`, a Rows component's is a typed list. If another component uses the same input name with another type the export reports the clash, as it does today.
- Ports are the built-in catalog's. A component imported from a module (`cuppa.component.json`) can declare a `port` per property in a later version of that format.
- Pack components are not changed (their properties are the pack's own).
