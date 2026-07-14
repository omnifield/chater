module github.com/omnifield/chater

go 1.26.0

// Toolchain pin (Go canon): builds fetch exactly this toolchain regardless of
// the host `go` version, so local/CI/devbox stay identical.
toolchain go1.26.5
