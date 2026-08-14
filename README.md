# binconv

[![CI](https://github.com/f0xtek/binconv/actions/workflows/ci.yml/badge.svg)](https://github.com/f0xtek/binconv/actions/workflows/ci.yml)

A small command-line tool for converting numbers between binary, octal, decimal and hexadecimal. Written in Go with no third-party dependencies.

```
$ binconv decimal hexadecimal 1337
0x539
```

## Features

- Convert between binary, octal, decimal and hexadecimal, in any direction
- Case-insensitive base names with short aliases (`bin`, `oct`, `dec`, `hex`, ...)
- Arbitrary-precision integers via `math/big` — no overflow, negatives supported
- Accepts `0x`/`0b`/`0o` prefixed input
- Convert multiple values in one call, or pipe values through stdin (one per line)
- Prefixed output (`0x539`, `0b1010`, `0o17`) by default, with an opt-out flag
- Clear exit codes for scripting: partial failures don't stop the batch

## Install

With Go installed:

```sh
go install github.com/f0xtek/binconv@latest
```

Or download a prebuilt binary from the [Releases](https://github.com/f0xtek/binconv/releases) page.

Or build from source (see [Development](#development) below).

## Usage

```
binconv [flags] <from> <to> [value ...]
... | binconv [flags] <from> <to>
```

`<from>` and `<to>` select the source and target base (see [Bases](#bases) below). Values can be
given as trailing arguments, or piped in on stdin (one value per line) when no values are given on
the command line.

### Examples

```sh
$ binconv decimal hexadecimal 1337
0x539

$ binconv dec hex --no-prefix 1337
539

$ binconv dec hex -u 1337
0x539

$ binconv hex dec 0x539 ff 10
1337
255
16

$ binconv hex dec -2a
-42

$ printf '1\n10\n11\n' | binconv bin dec
1
2
3

$ binconv dec hex 123456789012345678901234567890
0x18ee90ff6c373e0ee4e3f0ad2

$ binconv --help
```

## Bases

Base names are case-insensitive and accept the following aliases:

| Canonical     | Aliases              |
|---------------|-----------------------|
| `binary`      | `bin`, `b`, `2`        |
| `octal`       | `oct`, `o`, `8`        |
| `decimal`     | `dec`, `d`, `10`       |
| `hexadecimal` | `hex`, `x`, `h`, `16`  |

## Flags

| Flag                | Description                                                           |
|---------------------|------------------------------------------------------------------------|
| `--no-prefix`        | Omit the `0b`/`0o`/`0x` prefix (prefixes are on by default; no effect on decimal) |
| `-u`, `--upper`      | Uppercase hexadecimal digits                                          |
| `-h`, `--help`       | Show help                                                             |
| `-V`, `--version`    | Show version                                                          |

Flags may appear anywhere on the command line, including after the positional arguments.

## Exit codes

| Code | Meaning                                                              |
|------|------------------------------------------------------------------------|
| `0`  | All conversions succeeded                                             |
| `1`  | At least one value failed to convert (successful values still printed) |
| `2`  | Usage error — bad base name, missing arguments, unknown flag           |

## Development

Requires Go 1.25+.

```sh
make build   # build ./bin/binconv
make test    # run the test suite
make vet     # go vet
make fmt     # check gofmt formatting
make check   # vet + fmt + test
make run ARGS="dec hex 1337"
```

See the [Makefile](Makefile) for all available targets (`make help`).
