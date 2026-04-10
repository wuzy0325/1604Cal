$ErrorActionPreference = 'Stop'

function Invoke-CheckedCommand {
    param(
        [string]$Command,
        [string[]]$Arguments
    )

    & $Command @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "command failed: $Command $($Arguments -join ' ')"
    }
}

Write-Host '[check] go test ./cmd/... ./internal/...'
Invoke-CheckedCommand -Command 'go' -Arguments @('test', './cmd/...', './internal/...')

Write-Host '[check] go vet ./cmd/... ./internal/...'
Invoke-CheckedCommand -Command 'go' -Arguments @('vet', './cmd/...', './internal/...')

Write-Host '[check] npm --prefix web run typecheck'
Invoke-CheckedCommand -Command 'npm' -Arguments @('--prefix', 'web', 'run', 'typecheck')

Write-Host '[check] npm --prefix web run lint'
Invoke-CheckedCommand -Command 'npm' -Arguments @('--prefix', 'web', 'run', 'lint')

Write-Host '[check] npm --prefix web run test'
Invoke-CheckedCommand -Command 'npm' -Arguments @('--prefix', 'web', 'run', 'test')
