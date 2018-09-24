# Snitch

A simple tool that collects TODOs in the source code and reports them as GitHub issues.

## How it works

1. Snitch finds an unreported TODO,
2. Reports it to the GitHub as an issue,
3. Assigns the Issue number to the TODO marking it a reported,
4. Commits the reported TODO to the git repo,
5. Repeats the process until all of the unreported TODOs are reported.

After that you are supposed to push the new reported TODOs yourself.

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

Make sure to enable full access to private repos. For some reason it's required to post issues.

[personal-token]: https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
