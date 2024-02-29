module github.com/kurtosis-tech/kurtosis/kurtosis_version

go 1.21

// NOTE: This module is a tiny module that contains ONLY a Go file (generated via Bash) which contains the current version of the repo, for use
// in the APIC and Engine servers, reporting their own version
// The generated file is Git-ignored (because otherwise generation of the file would affect the "-dirty" state of the repo)

require github.com/sirupsen/logrus v1.9.3

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
