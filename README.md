# Snitch

A simple tool that collects TODOs in the source code and reports them as GitHub issues.

## TODO Format

### Unreported TODO

```
// TODO: rewrite this in Rust
```

### Reported TODO

```
// TODO(#42): rewrite this in Rust
```

## Usage

```
$ go run main.go list     # lists all TODOs in the current dir
$ go run main.go report   # report all unreported TODOs in the current dir
```

## GitHub Credentials File

`~/.snitch/config.ini`
