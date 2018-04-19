# Developer Guide

This document describes setting up your environment, as well as installing Chargeback.

## Development Dependencies

- Go 1.8 or higher
- Helm CLI 2.6.2 or higher
- Make
- Docker
- Dep
- jq
- realpath
- python (2 or 3 is ***REMOVED***ne).
  - pyyaml

If you're using MacOS with homebrew you can install all of these using the
following:

```
$ brew install go kubernetes-helm make docker dep jq coreutils
$ brew install python@2
$ pip2 install pyyaml
```

## Go Dependencies

We use [dep](https://golang.github.io/dep/docs/introduction.html) for managing
dependencies.

Dep installs dependencies into the `vendor/` directory at the
root of the repository, and to ensure everyone is using the same dependencies,
and ensure that if dependencies disappear, we commit the contents of `vendor/`
into git.

### Adding new dependencies

To add a new dependencies, you can generally follow the dep documentation.
Start by reading [https://golang.github.io/dep/docs/daily-dep.html](https://golang.github.io/dep/docs/daily-dep.html)
and you it should cover the most common things you'll be using dep for.

Otherwise, you should be able to just add a new import, and run `make vendor`
and the dependency will be installed.

When committing new dependencies, please use the following guidelines:

- Always commit changes to dependencies separately from other changes.
- Use one commit for changes to `Gopkg.toml`, and another commit for changes to
  `Gopkg.lock` and `vendor`.' Commit messages should be in the following forms:
  - `Gopkg.toml: Add new dependency $your_new_dependency`
  - `Gopkg.lock,vendor: Add new dependency $your_new_dependency`
