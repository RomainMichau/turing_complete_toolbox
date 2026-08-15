# Turing Complete Toolbox

A tiny local web app with helpers for playing [Turing Complete](https://turingcomplete.game/).
All the logic lives in Go; the page is a thin client that posts each keystroke to the backend.

## Install

```sh
go install .        # -> $GOPATH/bin/turing_complete_toolbox (already on PATH)
```

The web assets are embedded, so the installed binary is self contained and can be
run from anywhere.

## Run

```sh
turing_complete_toolbox             # serves on :8080 and opens a browser tab
turing_complete_toolbox -addr :9000
turing_complete_toolbox -open=false # no browser tab

go run .                            # from the source tree, without installing
```

## Tools

| Tool | What it does |
| --- | --- |
| Bits → Int | Binary string to unsigned / signed (two's complement) / hex / octal. Spaces, tabs and underscores are ignored. |

## Adding a tool

Drop a new file in `internal/tools/` and register it from `init()`:

```go
func init() {
    Register(&Tool{
        ID:          "my-tool",
        Name:        "My Tool",
        Description: "What it does.",
        Placeholder: "example input",
        Run: func(input string) ([]Field, error) {
            return []Field{{Label: "Result", Value: input}}, nil
        },
    })
}
```

It shows up automatically as a card in the UI — nothing to change in the frontend.

## Test

```sh
go test ./...
```
