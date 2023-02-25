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



