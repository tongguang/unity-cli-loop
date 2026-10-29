<#
Checks this project against the Tuanjie Editor that already has it open: the
Editor must be able to take a command, the project must compile with zero
errors and warnings, and a small EditMode test set must pass. Exits 1 on the
first failing step.

It never launches, restarts, or focuses the Editor. It stops when the Editor
is in Play Mode, because a compile would discard the Play session, and runs
the tests with --unsaved-changes fail so unsaved Scene changes stop the test
run instead of being saved.

The default tloop is this checkout's dist binary; rebuild it with
scripts/build-go-cli.sh after CLI source changes, or pass -TloopPath.

Usage:
  pwsh -NoProfile -File ./scripts/check-tuanjie.ps1
  pwsh -NoProfile -File ./scripts/check-tuanjie.ps1 -TestFilterType regex -TestFilterValue CliPin
  pwsh -NoProfile -File ./scripts/check-tuanjie.ps1 -SkipTests
  powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\check-tuanjie.ps1
#>

param(
    [string]$ProjectPath = "",
    [string]$TloopPath = "",
    [ValidateSet("exact", "regex", "assembly", "class")]
    [string]$TestFilterType = "class",
    [ValidateNotNullOrEmpty()]
    [string]$TestFilterValue = "CliVersionComparerTests",
    [ValidateRange(1, 1500)]
    [int]$TestTimeoutSeconds = 600,
    [switch]$SkipTests
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

[string]$RepoRoot = Split-Path -Parent $PSScriptRoot

function Exit-WithFailure {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Message,
        [string]$Details = ""
    )

    if (-not [string]::IsNullOrWhiteSpace($Details)) {
        Write-Host $Details
    }
    Write-Host "FAILED: $Message" -ForegroundColor Red
    exit 1
}

function Get-DefaultTloopPath {
    if ($env:OS -eq "Windows_NT") {
        return Join-Path $RepoRoot "dist/windows-amd64/tloop.exe"
    }

    [string]$os = if ((uname -s) -eq "Darwin") { "darwin" } else { "linux" }
    [string]$machine = uname -m
    [string]$arch = if ($machine -eq "arm64" -or $machine -eq "aarch64") { "arm64" } else { "amd64" }
    return Join-Path $RepoRoot "dist/$os-$arch/tloop"
}

function ConvertTo-WindowsPowerShellNativeArgument {
    param(
        [string]$Argument
    )

    # Windows argv rules treat backslashes specially only before a quote or the closing quote.
    # Do not double every backslash: ordinary paths such as C:\Users must stay unchanged.
    [string]$escapedArgument = [regex]::Replace($Argument, '(\\*)"', {
        param($match)

        [string]$backslashes = $match.Groups[1].Value
        return $backslashes + $backslashes + '\"'
    })

    if ($Argument -notmatch '\s') {
        return $escapedArgument
    }

    return [regex]::Replace($escapedArgument, '(\\*)\z', {
        param($match)

        [string]$backslashes = $match.Groups[1].Value
        return $backslashes + $backslashes
    })
}

function Invoke-Tloop {
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$CommandArguments
    )

    [string[]]$arguments = @($CommandArguments) + @("--project-path", $ResolvedProjectPath)
    if ($PSVersionTable.PSVersion.Major -lt 6) {
        $arguments = @($arguments | ForEach-Object { ConvertTo-WindowsPowerShellNativeArgument -Argument $_ })
    }

    # Windows PowerShell can turn native stderr lines into error records; under
    # Stop the first one would end the script before the JSON on stdout and the
    # exit code are read.
    [string]$previousErrorActionPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        [object[]]$output = @(& $ResolvedTloopPath @arguments)
        [int]$exitCode = $LASTEXITCODE
    }
    finally {
        $ErrorActionPreference = $previousErrorActionPreference
    }

    return [pscustomobject]@{
        ExitCode = $exitCode
        Text = (@($output | ForEach-Object { $_.ToString() }) -join [Environment]::NewLine)
    }
}

function ConvertFrom-TloopJson {
    param(
        [Parameter(Mandatory = $true)]
        [pscustomobject]$Result,
        [Parameter(Mandatory = $true)]
        [string]$CommandName
    )

    if ([string]::IsNullOrWhiteSpace($Result.Text)) {
        Exit-WithFailure -Message "tloop $CommandName printed nothing on stdout (exit code $($Result.ExitCode)); see the error above."
    }

    try {
        return $Result.Text | ConvertFrom-Json
    }
    catch {
        Exit-WithFailure -Message "tloop $CommandName did not print JSON (exit code $($Result.ExitCode))." -Details $Result.Text
    }
}

function Get-JsonValue {
    param(
        [Parameter(Mandatory = $true)]
        [object]$Json,
        [Parameter(Mandatory = $true)]
        [string]$Name
    )

    $property = $Json.PSObject.Properties[$Name]
    if ($null -eq $property) {
        return $null
    }
    return $property.Value
}

function Write-Step {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name
    )

    Write-Host ""
    Write-Host "=== $Name ==="
}

function Invoke-StatusStep {
    Write-Step -Name "Editor status"
    [pscustomobject]$result = Invoke-Tloop -CommandArguments @("status")
    $status = ConvertFrom-TloopJson -Result $result -CommandName "status"
    if ((Get-JsonValue -Json $status -Name "Ready") -ne $true) {
        Exit-WithFailure -Message "The Editor cannot take a command now (State: $(Get-JsonValue -Json $status -Name 'State'))." -Details $result.Text
    }
    if ((Get-JsonValue -Json $status -Name "IsPlaying") -eq $true) {
        Exit-WithFailure -Message "The Editor is in Play Mode. Stop it first: a compile would discard the Play session."
    }
    Write-Host "Ready"
}

function Invoke-CompileStep {
    Write-Step -Name "Compile"
    [pscustomobject]$result = Invoke-Tloop -CommandArguments @("compile")
    $compile = ConvertFrom-TloopJson -Result $result -CommandName "compile"
    $errorCount = Get-JsonValue -Json $compile -Name "ErrorCount"
    $warningCount = Get-JsonValue -Json $compile -Name "WarningCount"
    if ((Get-JsonValue -Json $compile -Name "Success") -ne $true -or $errorCount -ne 0 -or $null -eq $warningCount) {
        Exit-WithFailure -Message "Compile did not succeed." -Details $result.Text
    }
    if ($warningCount -ne 0) {
        Exit-WithFailure -Message "Compile reported $warningCount warning(s)." -Details $result.Text
    }
    Write-Host "0 errors, 0 warnings"
}

function Invoke-TestStep {
    Write-Step -Name "EditMode tests ($TestFilterType $TestFilterValue)"
    [pscustomobject]$result = Invoke-Tloop -CommandArguments @(
        "run-tests",
        "--test-mode", "EditMode",
        "--filter-type", $TestFilterType.ToLowerInvariant(),
        "--filter-value", $TestFilterValue,
        "--unsaved-changes", "fail",
        "--timeout-seconds", [string]$TestTimeoutSeconds
    )
    $tests = ConvertFrom-TloopJson -Result $result -CommandName "run-tests"
    if ((Get-JsonValue -Json $tests -Name "NoTestsFound") -eq $true) {
        Exit-WithFailure -Message "No EditMode test matched --filter-type $TestFilterType --filter-value $TestFilterValue." -Details $result.Text
    }
    if ((Get-JsonValue -Json $tests -Name "Success") -ne $true) {
        Exit-WithFailure -Message "EditMode tests did not pass (Status: $(Get-JsonValue -Json $tests -Name 'Status'))." -Details $result.Text
    }
    Write-Host "$(Get-JsonValue -Json $tests -Name 'PassedCount') passed, $(Get-JsonValue -Json $tests -Name 'SkippedCount') skipped"
}

[string]$ResolvedProjectPath = if ([string]::IsNullOrWhiteSpace($ProjectPath)) {
    $RepoRoot
}
else {
    (Resolve-Path -LiteralPath $ProjectPath).Path
}

if ([string]::IsNullOrWhiteSpace($TloopPath)) {
    $TloopPath = Get-DefaultTloopPath
}
if (-not (Test-Path -LiteralPath $TloopPath -PathType Leaf)) {
    Exit-WithFailure -Message "tloop was not found at $TloopPath. Build it with scripts/build-go-cli.sh or pass -TloopPath."
}
[string]$ResolvedTloopPath = (Resolve-Path -LiteralPath $TloopPath).Path

# tloop writes UTF-8, but PowerShell decodes native output with the console code
# page, a legacy one such as 936 on many Windows systems. That garbles non-ASCII
# paths and can swallow the path separator after them, which breaks the JSON.
[System.Text.Encoding]$previousOutputEncoding = [Console]::OutputEncoding
[Console]::OutputEncoding = New-Object System.Text.UTF8Encoding $false
try {
    Invoke-StatusStep
    Invoke-CompileStep
    if ($SkipTests) {
        Write-Step -Name "EditMode tests"
        Write-Host "Skipped (-SkipTests)"
    }
    else {
        Invoke-TestStep
    }
}
finally {
    [Console]::OutputEncoding = $previousOutputEncoding
}

Write-Host ""
Write-Host "Tuanjie check passed." -ForegroundColor Green
exit 0
