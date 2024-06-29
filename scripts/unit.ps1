$exitCode = 0

# Run unit tests
Write-Host "--> Running unit tests"
$packages = go list ./... |  Where-Object {$_ -inotmatch 'vendor'} | Where-Object {$_ -inotmatch 'examples'}

foreach ($package in $packages) {
  Write-Host "Running tests for $package"
  go test -race -v $package --test.skip TestHandleBasedMessageTestsWithBinary
  if ($LastExitCode -ne 0) {
    Write-Host "ERROR: Test failed, logging failure"
    $exitCode=1
  }
}



# go test -v -race github.com/pact-foundation/pact-go/v2
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2"
#   $exitCode=1
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/command
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/command"
#   $exitCode=1
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/consumer
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/consumer"
#   $exitCode=1
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/installer
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/installer"
#   $exitCode=1
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/internal/checker
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/internal/checker"
#   $exitCode
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/internal/native
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/internal/native"
#   $exitCode
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/internal/native/io.pact.plugin
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/internal/native/io.pact.plugin"
#   $exitCode
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/log
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/log"
#   $exitCode
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/matchers
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/matchers"
#   $exitCode
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/message
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/message"
#   $exitCode
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/message/v3
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/message/v3"
#   $exitCode
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/message/v4
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/message/v4"
#   $exitCode
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/models
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/models"
#   $exitCode
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/provider
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/provider"
#   $exitCode
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/proxy
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/proxy"
#   $exitCode
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/utils
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/utils"
#   $exitCode
# }
# go test -v -race github.com/pact-foundation/pact-go/v2/version
# if ($LastExitCode -ne 0) {
#   Write-Host "ERROR: Test failed, logging failure for pact-go/v2/version"
#   $exitCode
# }

Write-Host "Done!"
if ($exitCode -ne 0) {
  Write-Host "--> Build failed, exiting"
  Exit $exitCode
}