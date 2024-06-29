$exitCode = 0
Write-Host "Running provider tests"
$env:SKIP_PUBLISH='true'
go test -v -race -timeout=30s -tags=provider -count=1 github.com/pact-foundation/pact-go/v2/examples/...
if ($LastExitCode -ne 0) {
  Write-Host "ERROR: Test failed, logging failure"
  $exitCode=1
}

Write-Host "Done!"
if ($exitCode -ne 0) {
  Write-Host "--> Build failed, exiting"
  Exit $exitCode
}