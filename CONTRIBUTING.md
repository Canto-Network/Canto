# Basic Guidelines

Canto is written in Go.  We prefer to use the latest version of go because this will prevent mixed runtimes, and thus, errors. 

## Dev Env setup
* go v1.20.0
  * golangci-lint
* Visual Studio Code
* Mac or Linux, no Windows

```bash
git clone https://github.com/Canto-Network/Canto
cd Canto
go install ./...
code .
```

## Branching & Releases

* Each major version of canto should change the module path in go.mod.  
* branches should be created for each non-state-breaking release, eg release/v5.0.x

### Pull Request Templates

There are three PR templates. The [default template](./.github/PULL_REQUEST_TEMPLATE.md) is for types `fix`, `feat`, and `refactor`. We also have a [docs template](./.github/PULL_REQUEST_TEMPLATE/docs.md) for documentation changes and an [other template](./.github/PULL_REQUEST_TEMPLATE/other.md) for changes that do not affect production code. When previewing a PR before it has been opened, you can change the template by adding one of the following parameters to the url:

* `template=docs.md`
* `template=other.md`

### PR Targeting

Ensure that you base and target your PR on the `main` branch.

All feature additions and all bug fixes must be targeted against `main`. Exception is for bug fixes which are only related to a released version. In that case, the related bug fix PRs must target against the release branch.
