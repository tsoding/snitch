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
$ snitch list     # lists all TODOs in the current dir
$ snitch report   # report all unreported TODOs in the current dir
```

## GitHub Credentials File

`~/.snitch/config.ini`
