# Developer documentation

## Tooling

* Docker
* Java (>= 19) - required for the Avro plugin example

## Key Branches

### `1.x.x` 

The previous major version. Only bug fixes and security updates will be considered.

### `master`

The `2.x.x` release line. Current major version.

## windows todo

1. avro plugin fails to start
2. plugins fail to cleanly shutdown which causes tests to hang and zombie plugin processes
  - Killing via task manager / taskkill outside of go process works
3. 

```sh
# failing
go test -v -race github.com/pact-foundation/pact-go/v2/consumer
go test -v -race github.com/pact-foundation/pact-go/v2/internal/native --test.skip TestHandleBasedMessageTestsWithBinary
go test -v -race github.com/pact-foundation/pact-go/v2/internal/native --test.run TestHandleBasedMessageTestsWithBinary
#  *** Test I/O incomplete 6s after exiting.
#  exec: WaitDelay expired before I/O complete
go test -v -race github.com/pact-foundation/pact-go/v2/message/v4

#  passing
go test -v -race github.com/pact-foundation/pact-go/v2/command
go test -v -race github.com/pact-foundation/pact-go/v2/installer
go test -v -race github.com/pact-foundation/pact-go/v2/matchers
go test -v -race github.com/pact-foundation/pact-go/v2/provider
go test -v -race github.com/pact-foundation/pact-go/v2/proxy
go test -v -race github.com/pact-foundation/pact-go/v2/utils
```

```
=== RUN   TestHandleBasedMessageTestsWithBinary
memory allocation of 824633720898 bytes failed
FAIL    github.com/pact-foundation/pact-go/v2/internal/native   1.015s
```