# charm-lint

Static analysis for the terminal pitfalls that break Charm-style TUIs: padding counted in bytes, UTF-8 sliced in half, byte offsets used as rune indexes, map iteration leaking into rendered output and command failures that vanish without a trace.

## Install

```bash
go install github.com/junhinhow/charm-lint/cmd/charm-lint@latest
```

As a library:

```bash
go get github.com/junhinhow/charm-lint
```

## Usage

Run every rule over the current module:

```bash
charm-lint ./...
```

The process exits with status 3 when it finds problems (the usual `go/analysis` convention), so it can gate CI.

### As a vet tool

```bash
go vet -vettool=$(which charm-lint) ./...
```

### From Go

```go
import (
	"github.com/junhinhow/charm-lint/lint"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(lint.Analyzers()...)
}
```

## Rules

| Rule | Detects |
| --- | --- |
| `bytelength-padding` | `strings.Repeat` padding measured with `len` or `utf8.RuneCountInString` |
| `byteslice` | `s[:n]` / `s[n:]` with a variable bound flowing into rendered output |
| `range-index-arith` | arithmetic with the index of `for i := range s` when `s` is a string |
| `nondeterministic-render` | `range` over a map inside `Render*`, `View*` or `String` |
| `silent-exec` | `exec.Command(...).Run/Output/CombinedOutput` with a discarded error or no `cmd.Stderr` |
| `discarded-stat-error` | `fi, _ := os.Stdin.Stat()` followed by `fi.Mode()` / `fi.Size()` |

### bytelength-padding

`len` counts bytes, so `ação` is 6 bytes but 4 cells wide. `utf8.RuneCountInString` counts runes, but CJK and emoji occupy two cells.

```go
// wrong: the border is wider than the label
border := strings.Repeat("─", len(label))

// right: measure display width
border := strings.Repeat("─", max(0, 20-ansi.StringWidth(label)))
```

### byteslice

Cutting a string by byte can split a rune and produce invalid UTF-8, which the terminal shows as `�`.

```go
// wrong
fmt.Println(s[:width])

// right
fmt.Println(ansi.Truncate(s, width, "…"))
```

The rule follows simple assignments and returns (within the same function) and flags slices whose bounds are not compile-time constants.

### range-index-arith

In `for i := range s` with a string, `i` is a byte offset.

```go
// wrong: i+1 may point inside a multi-byte rune
for i := range s {
	next := s[i+1]
	_ = next
}

// right
for _, r := range s {
	_ = r
}
```

### nondeterministic-render

Go randomizes map iteration order. Rendering a map directly makes snapshots and tests flaky.

```go
// wrong
for k, v := range m {
	fmt.Fprintf(&b, "%s=%d\n", k, v)
}

// right
for _, k := range slices.Sorted(maps.Keys(m)) {
	fmt.Fprintf(&b, "%s=%d\n", k, m[k])
}
```

### silent-exec

A command that fails without its error or stderr surfacing leaves the UI apparently working.

```go
// wrong: any failure disappears
_ = exec.Command("git", "status").Run()

// right
cmd := exec.Command("git", "status")
cmd.Stderr = os.Stderr
if err := cmd.Run(); err != nil {
	return fmt.Errorf("git status: %w", err)
}
```

### discarded-stat-error

`Stat` returns a nil `FileInfo` when it fails, so using it afterwards panics.

```go
// wrong: nil pointer dereference when Stat fails
info, _ := os.Stdin.Stat()
width := info.Size()

// right
info, err := os.Stdin.Stat()
if err != nil {
	return err
}
width := info.Size()
```

## API

| Symbol | Description |
| --- | --- |
| `Analyzers() []*analysis.Analyzer` | All six rules, ready for `unitchecker` |
| `Charmlint` | Umbrella analyzer used by the binary; runs every rule on the same pass |
| `ByteLengthPadding` | `strings.Repeat` padding measured in bytes or runes |
| `ByteSlice` | String slicing by byte that reaches rendered output |
| `RangeIndexArith` | Byte offset from a string range used in arithmetic |
| `NondeterministicRender` | Map iteration inside a render function |
| `SilentExec` | External command without error handling or `cmd.Stderr` |
| `DiscardedStatError` | `os.Stdin/Stdout/Stderr.Stat` error discarded before `Mode`/`Size` |

## Using in CI

GitHub Actions:

```yaml
- name: charm-lint
  run: |
    go install github.com/junhinhow/charm-lint/cmd/charm-lint@latest
    charm-lint ./...
```

Or through `go vet`, which fits an existing vet step:

```bash
go install github.com/junhinhow/charm-lint/cmd/charm-lint@latest
go vet -vettool=$(which charm-lint) ./...
```

## Design notes

- Pure `go/analysis`; no `SuggestedFix` yet. Every diagnostic names the API that fixes the problem, so the correction is explicit in the message.
- `byteslice` is intraprocedural: it follows parentheses, string concatenation, assignments and returns within the same file, up to a small depth. Values that cross function boundaries are not tracked.
- `nondeterministic-render` only reports map loops that actually produce output (Print/Sprintf/Render/Write/Format calls, string concatenation or `+=`). A loop that only collects keys for later sorting is not reported; a loop that only appends values to a slice without formatting is a known false negative.
- `silent-exec` checks for `cmd.Stderr` at function scope, not per command. `CombinedOutput` still triggers the stderr check, as specified by the rule.
- Test files (`_test.go`) are analyzed like any other file.
- Analyzer names cannot contain hyphens (`analysis.Validate` requires identifiers), so the hyphenated rule IDs live in `Diagnostic.Category` and in this table.
- The tool needs the target packages to type-check; generated code and build-tag-specific files follow the usual `go/packages` rules.

## License

MIT
