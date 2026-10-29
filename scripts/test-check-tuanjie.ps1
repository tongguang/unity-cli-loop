<#
Verifies scripts/check-tuanjie.ps1 against a fake tloop that prints canned JSON:
which tloop commands each outcome runs, the exit code, the arguments passed to
run-tests, and that UTF-8 output with non-ASCII paths still parses under a
legacy console code page. Windows only, because the fake tloop is a batch file.
#>

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if ($env:OS -ne "Windows_NT") {
    Write-Host "Skipping check-tuanjie tests: the fake tloop is a Windows batch file."
    exit 0
}

[string]$RepoRoot = Split-Path -Parent $PSScriptRoot
[string]$CheckScriptPath = Join-Path $RepoRoot "scripts/check-tuanjie.ps1"
[System.Text.Encoding]$Utf8NoBom = New-Object System.Text.UTF8Encoding $false
# Decoded with this double-byte code page, the three UTF-8 encoded CJK characters
# below end on a lead byte that swallows the path separator after them.
[int]$LegacyCodePage = 936
[string]$NonAsciiFolder = -join @([char]0x9879, [char]0x76EE, [char]0x7EC4)
[string]$NonAsciiProjectRootJson = 'F:\\' + $NonAsciiFolder + '\\Game'

[string]$FakeTloopScript = @'
@echo off
setlocal
set "DIR=%~dp0"
>>"%DIR%calls.log" echo %*
set "CODE=0"
if exist "%DIR%%~1.exitcode" set /p CODE=<"%DIR%%~1.exitcode"
if exist "%DIR%%~1.json" (
  type "%DIR%%~1.json"
) else (
  echo fake tloop has no response for %~1 1>&2
  set "CODE=1"
)
exit /b %CODE%
'@

[string]$ReadyStatus = '{"State":"Ready","Ready":true,"IsPlaying":false,"NextActions":[]}'
[string]$CleanCompile = '{"Success":true,"ErrorCount":0,"WarningCount":0,"ProjectRoot":"' + $NonAsciiProjectRootJson + '"}'
[string]$PassedTests = '{"Success":true,"Status":"Passed","NoTestsFound":false,"TestCount":2,"PassedCount":2,"SkippedCount":0}'

function Invoke-Scenario {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,
        [hashtable]$Responses = @{},
        [hashtable]$ExitCodes = @{},
        [hashtable]$Parameters = @{},
        [Parameter(Mandatory = $true)]
        [int]$ExpectedExitCode,
        [string[]]$ExpectedCommands = @(),
        [string[]]$ExpectedOutput = @(),
        [string[]]$ExpectedArguments = @()
    )

    [string]$scenarioRoot = Join-Path $TestRoot $Name
    [string]$projectDir = Join-Path $scenarioRoot "project"
    [string]$binDir = Join-Path $scenarioRoot "bin"
    New-Item -ItemType Directory -Path $projectDir, $binDir -Force | Out-Null
    [string]$fakeTloopPath = Join-Path $binDir "tloop.cmd"
    [System.IO.File]::WriteAllText($fakeTloopPath, ($FakeTloopScript -replace "`r?`n", "`r`n"), [System.Text.Encoding]::ASCII)
    foreach ($command in $Responses.Keys) {
        [System.IO.File]::WriteAllText((Join-Path $binDir "$command.json"), $Responses[$command], $Utf8NoBom)
    }
    foreach ($command in $ExitCodes.Keys) {
        [System.IO.File]::WriteAllText((Join-Path $binDir "$command.exitcode"), [string]$ExitCodes[$command], [System.Text.Encoding]::ASCII)
    }

    [hashtable]$checkParameters = @{ ProjectPath = $projectDir; TloopPath = $fakeTloopPath }
    foreach ($key in $Parameters.Keys) {
        $checkParameters[$key] = $Parameters[$key]
    }

    [System.Text.Encoding]$originalEncoding = [Console]::OutputEncoding
    [Console]::OutputEncoding = [System.Text.Encoding]::GetEncoding($LegacyCodePage)
    try {
        [string]$output = & $CheckScriptPath @checkParameters *>&1 | Out-String
        [int]$exitCode = $LASTEXITCODE
        [int]$codePageAfterRun = [Console]::OutputEncoding.CodePage
    }
    finally {
        [Console]::OutputEncoding = $originalEncoding
    }

    if ($exitCode -ne $ExpectedExitCode) {
        throw "${Name}: expected exit code $ExpectedExitCode, got $exitCode.`n$output"
    }
    if ($codePageAfterRun -ne $LegacyCodePage) {
        throw "${Name}: the console code page was not restored (got $codePageAfterRun)."
    }

    [string]$callsLogPath = Join-Path $binDir "calls.log"
    [string[]]$callLines = @()
    if (Test-Path -LiteralPath $callsLogPath) {
        $callLines = @(Get-Content -LiteralPath $callsLogPath)
    }
    [string]$actualCommands = @($callLines | ForEach-Object { ($_ -split ' ')[0] }) -join ','
    [string]$expectedCommandList = $ExpectedCommands -join ','
    if ($actualCommands -ne $expectedCommandList) {
        throw "${Name}: expected tloop commands [$expectedCommandList], got [$actualCommands].`n$output"
    }

    foreach ($fragment in $ExpectedOutput) {
        if (-not $output.Contains($fragment)) {
            throw "${Name}: expected the output to contain [$fragment].`n$output"
        }
    }
    [string]$callsText = $callLines -join "`n"
    foreach ($fragment in $ExpectedArguments) {
        if (-not $callsText.Contains($fragment)) {
            throw "${Name}: expected the tloop arguments to contain [$fragment].`n$callsText"
        }
    }

    Write-Host "ok - $Name"
}

[string]$TestRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("tloop-check-tuanjie-" + [Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $TestRoot | Out-Null
try {
    # Verifies that a ready Editor, a clean compile, and passing tests exit 0, that UTF-8 JSON with a
    # non-ASCII path parses under a legacy code page, and that run-tests gets a lowercase filter type
    # and --unsaved-changes fail.
    Invoke-Scenario -Name "passes-when-every-step-succeeds" `
        -Responses @{ "status" = $ReadyStatus; "compile" = $CleanCompile; "run-tests" = $PassedTests } `
        -Parameters @{ TestFilterType = "Regex"; TestFilterValue = "CliPin" } `
        -ExpectedExitCode 0 `
        -ExpectedCommands @("status", "compile", "run-tests") `
        -ExpectedOutput @("Tuanjie check passed.") `
        -ExpectedArguments @("--test-mode EditMode", "--filter-type regex --filter-value CliPin", "--unsaved-changes fail")

    # Verifies that an Editor that cannot take a command stops the check before compiling.
    Invoke-Scenario -Name "stops-when-editor-is-not-ready" `
        -Responses @{ "status" = '{"State":"NotRunning","Ready":false,"NextActions":["Start Tuanjie."]}' } `
        -ExitCodes @{ "status" = 1 } `
        -ExpectedExitCode 1 `
        -ExpectedCommands @("status") `
        -ExpectedOutput @("State: NotRunning")

    # Verifies that Play Mode stops the check before a compile could discard the Play session.
    Invoke-Scenario -Name "stops-in-play-mode" `
        -Responses @{ "status" = '{"State":"Ready","Ready":true,"IsPlaying":true}' } `
        -ExpectedExitCode 1 `
        -ExpectedCommands @("status") `
        -ExpectedOutput @("Play Mode")

    # Verifies that compile errors fail the check on a non-zero tloop exit and print the response
    # with its non-ASCII path intact.
    Invoke-Scenario -Name "fails-on-compile-errors" `
        -Responses @{
            "status" = $ReadyStatus
            "compile" = '{"Success":false,"ErrorCount":1,"WarningCount":0,"Errors":[{"File":"' + $NonAsciiProjectRootJson + '\\Assets\\Broken.cs","Message":"error CS1002: ; expected"}]}'
        } `
        -ExitCodes @{ "compile" = 1 } `
        -ExpectedExitCode 1 `
        -ExpectedCommands @("status", "compile") `
        -ExpectedOutput @("Compile did not succeed.", $NonAsciiFolder)

    # Verifies that compile warnings fail the check before any test runs.
    Invoke-Scenario -Name "fails-on-compile-warnings" `
        -Responses @{ "status" = $ReadyStatus; "compile" = '{"Success":true,"ErrorCount":0,"WarningCount":2}' } `
        -ExpectedExitCode 1 `
        -ExpectedCommands @("status", "compile") `
        -ExpectedOutput @("2 warning(s)")

    # Verifies that a filter matching no test fails the check instead of passing with zero tests.
    Invoke-Scenario -Name "fails-when-no-test-matches" `
        -Responses @{
            "status" = $ReadyStatus
            "compile" = $CleanCompile
            "run-tests" = '{"Success":true,"Status":"NoTestsFound","NoTestsFound":true,"TestCount":0}'
        } `
        -ExpectedExitCode 1 `
        -ExpectedCommands @("status", "compile", "run-tests") `
        -ExpectedOutput @("No EditMode test matched")

    # Verifies that a failed test run fails the check and reports its status.
    Invoke-Scenario -Name "fails-on-test-failures" `
        -Responses @{
            "status" = $ReadyStatus
            "compile" = $CleanCompile
            "run-tests" = '{"Success":false,"Status":"Failed","NoTestsFound":false,"TestCount":2,"FailedCount":1}'
        } `
        -ExitCodes @{ "run-tests" = 1 } `
        -ExpectedExitCode 1 `
        -ExpectedCommands @("status", "compile", "run-tests") `
        -ExpectedOutput @("Status: Failed")

    # Verifies that -SkipTests stops after the compile and still exits 0.
    Invoke-Scenario -Name "skip-tests-runs-no-tests" `
        -Responses @{ "status" = $ReadyStatus; "compile" = $CleanCompile } `
        -Parameters @{ SkipTests = $true } `
        -ExpectedExitCode 0 `
        -ExpectedCommands @("status", "compile") `
        -ExpectedOutput @("Skipped (-SkipTests)", "Tuanjie check passed.")

    # Verifies that output that is not JSON fails the check with the raw text.
    Invoke-Scenario -Name "fails-on-non-json-output" `
        -Responses @{ "status" = "panic: runtime error" } `
        -ExitCodes @{ "status" = 2 } `
        -ExpectedExitCode 1 `
        -ExpectedCommands @("status") `
        -ExpectedOutput @("did not print JSON", "panic: runtime error")

    # Verifies that a missing tloop binary fails the check before any command runs.
    Invoke-Scenario -Name "fails-when-tloop-is-missing" `
        -Parameters @{ TloopPath = (Join-Path $TestRoot "missing\tloop.exe") } `
        -ExpectedExitCode 1 `
        -ExpectedCommands @() `
        -ExpectedOutput @("tloop was not found")
}
finally {
    Remove-Item -LiteralPath $TestRoot -Recurse -Force
}

Write-Host "check-tuanjie tests passed."
exit 0
