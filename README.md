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

Resides in `$HOME/.snitch/config.ini`

## Format

```ini
[github]
personal_token = <personal-token>
```

Checkout [GitHub Help][personal-token] on how to get the Personal Access Token.

[personal-token]: https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
