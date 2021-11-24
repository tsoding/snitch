[![Tsoding](https://img.shields.io/badge/twitch.tv-tsoding-purple?logo=twitch&style=for-the-badge)](https://www.twitch.tv/tsoding)
[![Build Status](https://travis-ci.org/tsoding/snitch.svg?branch=master)](https://travis-ci.org/tsoding/snitch)
[![Build Status](https://github.com/tsoding/snitch/workflows/Go/badge.svg)](https://github.com/tsoding/snitch/actions?query=workflow%3AGo)

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

#### Example

```
// TODO: rewrite this in Rust
```

#### Parsing

Regular expression: `^(.*)TODO(O*): (.*)$` [Play](https://regex101.com/r/u5lkxf/2)

Capture Groups:
- Group 1: **Prefix**. Used only to precisly recover the text of the line where the TODO is originally located.
- Group 2: **Urgency Suffix**. Used to indicate the urgency of the TODO. The higher the amount of `O`-s, the more urgent the TODO is. (See [Urgency](#Urgency) for more info)
- Group 3: **Suffix**. Used as the title of the issue.

### Reported TODO

#### Example

```
// TODO(#42): rewrite this in Rust
```

#### Parsing

Regular expression: `^(.*)TODO(O*)\((.*)\): (.*)$` [Play](https://regex101.com/r/5U6rjS/1)

Capture Groups:
- Group 1: **Prefix**. Used only to precisly recover the text of the line where the TODO is originally located.
- Group 2: **Urgency Suffix**. Used to indicate the urgency of the TODO. The higher the amount of `O`-s, the more urgent the TODO is. (See [Urgency](#Urgency) for more info)
- Group 3: **ID**. The number of the Issue.
- Group 4: **Suffix**. Used as the title of the issue.

### TODO Body

#### Example

```
// TODO: rewrite this in Rust
//   I honestly think Rust is going to be around forever,
//   I really do. I think this is like, this is the formation
//   of Ancient Greek.
//   © https://www.reddit.com/r/programmingcirclejerk/comments/ahmnwa/i_honestly_think_rust_is_going_to_be_around/
```

#### Parsing

- Snitch remembers the TODO's prefix.
- Snitch parses all of the consecutive lines with the same prefix as the body.
- The body is reported as the Issue Description.

### Urgency

The urgency system was stolen from [fixmee](https://github.com/rolandwalker/fixmee#explanation) Emacs extension. The urgency of TODOs is indicated by repetitions of the final character of the keyword. For example, one might write TODOOOOOOOOO for an important issue. The `list` subcommand will sort the TODOs in the descending order by their urgency.

## Remote specification

By default Snitch will automatically reference the `origin` remote as the defacto standard for most projects.

However, you can specify which remote Snitch uses on a per repo basis.

### .snitch.yaml

Remotes are defined in `.snitch.yaml` under **remote**.

#### Example
<pre>
title:
  transforms:
    - match: (.*) \*/
      replace: $1
    - match: (.*) \*\}
      replace: $1
keywords:
  - FIXME
<b>remote: &ltremote&gt</b>
</pre>

### Commandline
Remotes can be dynamically specified when Snitch is called.

This will overide a previously defined remote in `.snitch.yaml`

#### Example
```
$ ./snitch report --remote <remote>
```

## Installation

```console
$ go get github.com/tsoding/snitch
```

## Credentials

### GitHub Credentials
Snitch obtains GitHub credentials from two places:  (default) environment variable or file.

#### Environment Variable
`export GITHUB_PERSONAL_TOKEN = <personal-token>` which can be added to `.bashrc`

#### File

Config file can be stored in one of the following directories:
- `$HOME/.config/snitch/github.ini`
- `$HOME/.snitch/github.ini`

Format:
```ini
[github]
personal_token = <personal-token>
```

Checkout [GitHub Help][personal-token] on how to get the Personal Access Token.

Make sure to enable full access to private repos. For some reason it's required to post issues.

### GitLab Credentials

GitLab credentials configuration is similar to GitHub with an exception that you also have to provide the host of the GitLab instance.

#### Environment Variable

`export GITLAB_PERSONAL_TOKEN = <personal-token>` which can be added to `.bashrc`.

Each of the credentials are to be separated by `,` and in format: `<host>:<personal-token>`. Credentials without host part, e.g. `<personal-token>` are interpreted as `gitlab.com` tokens to maintain backward compatibility and invalid tokens are ignored (prints an error message).

#### File

Config file can be stored in one of the following directories:
- `$HOME/.config/snitch/gitlab.ini`
- `$HOME/.snitch/gitlab.ini`

Format:

```ini
[gitlab.com]
personal_token = <personal-token>

[gitlab.local]
personal_token = <personal-token>
```

Checkout [GitLab Help][personal-token-gitlab] on how to get the Personal Access Token. Make sure to enable `api` scope for the token.

## Usage

For usage help just run `snitch` without any arguments:

```console
$ ./snitch
```

## .snitch.yaml

### Custom keywords

You don't have to use `TODO` as the keyword of a todo you want to
"snitch up". The keyword is customizable through `.snitch.yaml`
config:

```yaml
keywords:
  - TODO
  - FIXME
  - XXX
  - "@todo"
```

### Issue Title Transformation

You can apply project local issue title transformations. Create
`.snitch.yaml` file in the root of the project with the following
content:

```yaml
title:
  transforms:
    - match: (.*) \*/
      replace: $1
```

This feature is very useful for removing garbage from the Issue
Titles. Like `*/` at the end of C comments.

## Development

```console
$ export GOPATH=$PWD
$ go get .
$ go build .
```

### Run tests

```console
$ go test ./...
```

For a more detailed output:

```console
$ go test -v -cover ./...
```

## Support

You can support my work via

- Twitch channel: https://www.twitch.tv/subs/tsoding
- Patreon: https://www.patreon.com/tsoding

[personal-token]: https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
[personal-token-gitlab]: https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html
