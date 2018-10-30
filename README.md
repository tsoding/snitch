[![Build Status](https://travis-ci.org/tsoding/snitch.svg?branch=master)](https://travis-ci.org/tsoding/snitch)
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

## Installation

```console
$ go get github.com/tsoding/snitch
```

## Usage

```console
$ snitch list                             # lists all TODOs in the current dir
$ snitch report <owner/repo> [issue-body] # report all unreported TODOs in the current dir
$ snitch purge <owner/repo>               # remove TODOs that refer to closed issues
```

## GitHub Credentials File

Resides in `$HOME/.snitch/github.ini`

### Format

```ini
[github]
personal_token = <personal-token>
```

Checkout [GitHub Help][personal-token] on how to get the Personal Access Token.

Make sure to enable full access to private repos. For some reason it's required to post issues.

[personal-token]: https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
