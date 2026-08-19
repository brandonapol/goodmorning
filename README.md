# goodmorning

A terminal app that browses your photos and renders them as ANSI/ASCII art
right in the terminal.

## Status

**Phase 1**: local folders only. Google Photos, iCloud, and network drives
are planned as additional sources — see the source adapter interface in
`internal/source`.

## Usage

```sh
go run . --path ~/Pictures
```

Or configure one or more named sources in `~/.config/goodmorning-photos/config.json`:

```json
{
  "sources": [
    { "name": "Pictures", "path": "/home/you/Pictures" }
  ]
}
```

Then just run:

```sh
go run .
```

Navigate with arrow keys / `hjkl` / `enter`, go back with `esc`, quit with `q`.
