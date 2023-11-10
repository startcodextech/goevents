# GO Events

## Table of contents
* [Go Tools Installation Guide](#go-tools-installation-guide)
  * [BUF](#buf)
  * [protoc-gen-go](#protoc-gen-go)
  * [Mockery](#mockery)

![CQRS](https://i.ibb.co/Y0wCX43/1-q-Cy2-p3v-9sbag-Bpex1-Cr-A.webp "CQRS")


## Go Tools Installation Guide

This guide provides detailed instructions for installing essential tools for protocol buffer management and mock code generation in Go. It includes steps for installing buf, protoc-gen-go, and mockery.

### BUF

`buf` is a command-line tool that facilitates linting, formatting, and generating code from protocol buffer files.

Install buf
On macOS systems, use `brew` to install `buf`:
```shell
brew install buf
```
**Note:** For other operating systems, refer to the [official `buf` documentation](https://docs.buf.build/installation) for platform-specific installation instructions.

### protoc-gen-go

`protoc-gen-go` is a plugin for the `protoc` protocol buffer compiler. It generates Go code from `.proto` files.

Install protoc-gen-go
To install `protoc-gen-go`, run the following command:
```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```
Then, update your `PATH` to include the Go binary directory to ensure the `protoc-gen-go` plugin can be located by your system:
```shell
export PATH="$PATH:$(go env GOPATH)/bin"
```
**Note:** Consider adding the `export PATH` line to your shell's configuration file (such as `.bashrc` or `.zshrc`) to make this change permanent.

### Mockery
`mockery` is a tool that provides an automated way to generate mocks for Go interfaces, simplifying unit testing.

Install mockery

Install `mockery` using the following command:

```shell
go install github.com/vektra/mockery/v2@latest
```

**Troubleshooting Tips:**

- Ensure your Go version is up to date to prevent compatibility issues when installing `mockery`.
- Confirm that `$(go env GOPATH)/bin` is in your `PATH` so that `mockery` can be executed from any directory.

With these tools installed, you'll be equipped to handle protocol buffers and generate mock code in Go, which are crucial for an efficient development workflow and for writing automated tests.