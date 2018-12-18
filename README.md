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

Keep in mind that Snitch supports only single line TODOs right
now. For a possible multiline support track issue [#81].

### Unreported TODO

#### Example

```
// TODO: rewrite this in Rust
```

#### Parsing

Regular expression: `^(.*)TODO: (.*)$` [Play](https://regex101.com/r/u5lkxf/2)

Capture Groups:
- Group 1: **Prefix**. Used only to precisly recover the text of the line where the TODO is originally located.
- Group 2: **Suffix**. Used as the title of the issue.

### Reported TODO

#### Example

```
// TODO(#42): rewrite this in Rust
```

#### Parsing

Regular expression: `^(.*)TODO\((.*)\): (.*)$` [Play](https://regex101.com/r/5U6rjS/1)

Capture Groups:
- Group 1: **Prefix**. Used only to precisly recover the text of the line where the TODO is originally located.
- Group 2: **ID**. The number of the Issue.
- Group 3: **Suffix**. Used as the title of the issue.

## Installation

```console
$ go get github.com/tsoding/snitch
```

## Usage

```console
$ snitch list                         # lists all TODOs in the current dir
$ snitch report [--body <issue-body>] # report all unreported TODOs in the current dir
$ snitch purge <owner/repo>           # remove TODOs that refer to closed issues
```

## Issue Title Transformation

You can apply project local issue title transformations. Create
`.snich.yaml` file in the root of the project with the following
content:

```yaml
title:
  transforms:
    - match: (.*) \*/
      replace: $1
```

This feature is very useful for removing garbage from the Issue
Titles. Like `*/` at the end of C comments.

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
[#81]: https://github.com/tsoding/snitch/issues/81
