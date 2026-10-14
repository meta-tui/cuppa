# Bundled packs

Packs that ship inside Cuppa (`libs/catalog/bundled/packs/*.cupp`). To propose one, follow
[Get your pack bundled with Cuppa](../packs.md#6-get-your-pack-bundled-with-cuppa-pull-request).

| Pack id | Name | For | Components | By |
|---|---|---|---|---|
| `starter` | Starter | Small building blocks to start from | Card, Alert | Cuppa |
| `widgets` | Widgets | Everyday interface pieces | Badge, Key hint, Stat card, Breadcrumb, Sidebar menu, Command palette | Cuppa |
| `controls` | Controls | Choice and disclosure controls | Checkbox, Radio, Toggle, Accordion (open), Accordion (closed) | Cuppa |

The divider is no longer in the Widgets pack: it is the **Divider** component of Lip Gloss
(`cuppa.divider`), horizontal or vertical, in a single, double, heavy, dashed or dotted line, with an optional
label. A design that used the pack's divider keeps it, since a `.cuppa` file carries a copy.
