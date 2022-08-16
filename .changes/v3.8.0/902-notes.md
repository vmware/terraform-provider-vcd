* Bumps Go to 1.18 in `go.mod` as the minimum required version. [GH-902]
* Does `make fmt` which adjusts doc formatting for Go 1.19 [GH-902]
* Adjusts GitHub actions to use the latest code [GH-902]
* `staticcheck` switched version naming from `2021.1.2` to `v0.3.3` in downloads section. This PR
  also updates the code to fetch correct staticcheck. [GH-902]
* package `io/ioutil` is deprecated as of Go 1.16. `staticcheck` started complaining about usage of
  deprecated packages. As a result this PR switches packages to either `io` or `os` (still the same
  functions are used) [GH-902]