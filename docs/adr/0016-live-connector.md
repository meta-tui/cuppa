# ADR 0016: A live connection between a running app and the designer

Status: superseded by ADR 0017 (removed in #271)

## Context

A design is checked against sample values. The real data (MVD's engine, its
file lists, its progress) only appears when the exported program runs, so a
layout problem that depends on data (a name that does not fit, 200 rows, an
empty state) is found late, and fixing it means leaving the designer.

The wish is the PowerApps one: the logic and the data stay in the application;
Cuppa is where the interface is drawn, and it can be plugged into the running
application to show its real data and to drive it.

The pieces already exist (ADR 0007): a page has typed **inputs** (its `Props`)
and typed **events** (returned data, not callbacks). A live connection only
has to carry inputs one way and events the other.

## Decision

The running application **publishes the inputs of its pages**; Cuppa draws the
design with them and **sends back the events** a click or key raises. Cuppa
holds no application logic and no data source.

```
  your app (any language)                          Cuppa designer
  ┌─────────────────────┐   state: page + inputs   ┌──────────────────┐
  │ engine, logic, data │ ───────────────────────▶ │ canvas / preview │
  │  listens on         │ ◀─────────────────────── │ draws the design │
  │  127.0.0.1:<port>   │   event: click / key     │ with live values │
  └─────────────────────┘                          └──────────────────┘
```

### Who listens

The **application listens, Cuppa connects.** The app opens a TCP port on the
loopback address only when it is asked to (an environment variable or a flag
the app chooses), so a released build does nothing. Cuppa dials, and redials
when the app restarts. Several designers can watch one app.

### The protocol

Newline-delimited JSON over TCP: one JSON object per line, UTF-8, `type` says
what it is. No library is needed to speak it, which is what lets the app be
written in any language.

App to Cuppa:

```json
{"type":"hello","protocol":1,"app":"MVD","pages":["Run","Download"]}
{"type":"state","page":"Run","current":true,"inputs":{"Title":"Run","Width":80,"Rows":["a","b"]}}
```

Cuppa to app:

```json
{"type":"hello","protocol":1,"design":"mvd","pages":["Run","Download","Colours"]}
{"type":"event","page":"Run","name":"ColoursSave","x":3,"y":0}
```

- `inputs` uses the input **names of the design** (View, Variables lists them),
  with the value of the input's kind: text, number, yes/no, a list of text, or
  a list of rows for a Rows component. A name the design does not have is
  listed in the status bar and ignored; an input the app did not send keeps its
  sample value.
- `state` is a full snapshot of the page's inputs, sent when they change. It
  is never a diff, so a missed message is repaired by the next one.
- `current: true` says which page the app is showing; the designer follows it
  unless a page was chosen by hand (the pages bar says "live").
- `event` is what `<Page>Handle` returns today: the event's name, and the cell
  inside the component for a click (0, 0 for a key). The app answers with a new
  `state`; Cuppa does not change any value itself, so what is drawn is always
  what the app said.
- A message that is not understood is ignored. `protocol` changes only when a
  message changes meaning.

### What the app has to do

1. **Get the contract.** It is the design: `cuppa design contract <file>`
   prints each page's inputs (name, kind) and events (name) as JSON, so a
   program in any language can check itself against it. For Go, the export
   writes it as types (`RunProps`, `RunEvent`) as it does today.
2. **Listen**, when asked to, and write a `state` line per page when its
   values change, and once for each page after `hello`.
3. **Read** `event` lines and do what it does for a click or key today.

For Go the export also writes a small registry of the pages (name, props
type, how to decode an event), and `libs/connector` (headless, standard
library only) turns it into the three steps:

```go
// sketch of the intended use
c := connector.Listen(os.Getenv("CUPPA_CONNECT"), pages.Registry())
defer c.Close()

c.Publish(pages.Run, runProps)            // when the props change
// in Update:  case run.RunEvent: ...     // events arrive as the typed events
cmd := c.Wait()                           // a tea.Cmd that returns the next event
```

`libs/connector` is a convenience, not a requirement: the app may open the
socket and write the lines itself.

A program in another language does the same with the sockets it has:

```python
# sketch: publish one page, print events
import json, socket
srv = socket.create_server(("127.0.0.1", 7331)); conn, _ = srv.accept()
f = conn.makefile("rw")
f.write(json.dumps({"type": "hello", "protocol": 1, "app": "demo", "pages": ["Run"]}) + "\n")
f.write(json.dumps({"type": "state", "page": "Run", "current": True,
                    "inputs": {"Title": "Run", "Width": 80}}) + "\n"); f.flush()
for line in f:
    print(json.loads(line))          # {"type":"event","page":"Run","name":"..."}
```

### What you do in Cuppa

1. Open the design of the app (`.cuppa`).
2. **File, Connect…** and type the address (`127.0.0.1:7331`), or start with
   `cuppa-tui --connect 127.0.0.1:7331 design.cuppa`. The address is not saved
   in the design: it is a property of this run, not of the design.
3. The status bar says `live: MVD` and the canvas shows the app's values. Edit
   the layout as usual; the values keep arriving.
4. **View, Preview** runs the design as today, and a click or key is also sent
   to the app as an event; the app's next `state` is what is drawn.
5. **File, Disconnect** (or the app closing) goes back to the samples. Nothing
   in the design file changes because of a connection.

### Limits and safety

- Loopback only. The connector refuses an address that is not `127.0.0.1`,
  `::1` or `localhost` unless the app passes an explicit option; there is no
  authentication in this step, so it is a development tool, not a remote
  management channel.
- Values, never code. Cuppa executes nothing it receives: a `state` is data
  for the design's inputs and an `event` is data for the app.
- The design stays the source of the interface and the app the source of the
  values. Moving a component in the designer while connected does not touch
  the app until the design is exported again; "Check Go export" says if the
  exported code differs from the design.

## Consequences

- MVD (or any app) gets its real data in the designer without a second
  program, and its logic stays where it is, in any language.
- The inputs and events of a design become an interface that other tools can
  read (`design contract`), which also helps the agent instructions (#174).
- A new headless lib (`libs/connector`) and a new adapter in `cuppa-tui`; the
  document, canvas and render libs do not change, and the file format does not
  change.
- Not covered: remote or authenticated connections, recording and replaying a
  session, several apps at once in one designer.

## Steps

1. `libs/connector`: the messages, a listener and a client, tests with a fake
   app (no terminal needed).
2. `cuppa-tui`: Connect…, Disconnect, live values on the canvas, `--connect`.
3. Events from the preview to the app.
4. `cuppa design contract`, and the generated page registry for Go. Built: the
   command prints each page's inputs (name, kind, fields of rows and hints) and
   events as JSON; each exported page also has `<Page>Inputs(props)`, the props
   by the input names of the design, and `<Page>EventNamed(name, x, y)`, the
   typed event of a name (the registry the sketch above calls `pages`; it is
   plain functions, so it needs no library).
5. The user-guide section and the agent instructions; MVD adopts it first.
