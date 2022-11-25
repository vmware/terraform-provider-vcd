* Bump Go to 1.19 in `go.mod` as the minimum required version. [GH-902] [GH-916]
* Code documentation formatting is adjusted using Go 1.19 (`make fmt`) [GH-902]
* Adjust GitHub actions in pipeline to use the latest code [GH-902]
* `staticcheck` switched version naming from `2021.1.2` to `v0.3.3` in downloads section. This PR
  also updates the code to fetch correct staticcheck. [GH-902]
* package `io/ioutil` is deprecated as of Go 1.16. `staticcheck` started complaining about usage of
  deprecated packages. As a result this PR switches packages to either `io` or `os` (still the same
  functions are used) [GH-902]
