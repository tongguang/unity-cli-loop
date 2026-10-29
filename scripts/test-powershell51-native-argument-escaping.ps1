<#
Verifies that Windows PowerShell native argument escaping preserves backslashes
while protecting embedded quotation marks for execute-dynamic-code requests.
#>

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# A non-ASCII byte parses on an English system but can break Windows PowerShell 5.1 on a CJK code
# page, so the check has to run here rather than rely on the host that happens to run CI.
[byte[]]$scriptBytes = [System.IO.File]::ReadAllBytes($PSCommandPath)
if (@($scriptBytes | Where-Object { $_ -gt 127 }).Count -gt 0) {
    throw "$PSCommandPath must stay ASCII so Windows PowerShell 5.1 parses it under any code page."
}

[string[]]$ScriptPaths = @(
    "scripts/run-windows-e2e.ps1",
    "scripts/check-tuanjie.ps1",
    "scripts/test-simulate-mouse-demo.ps1",
    "scripts/test-screenshot-uitoolkit-annotation.ps1"
    # Disabled along with input recording and replay.
    # "Assets/Tests/Demo/scripts/verify-replay-via-cli.ps1"
)

function Get-FunctionDefinition {
    param(
        [string]$ScriptPath,
        [string]$FunctionName
    )

    [string]$source = Get-Content -LiteralPath $ScriptPath -Raw -Encoding UTF8
    [string]$marker = "function $FunctionName"
    [int]$start = $source.IndexOf($marker, [System.StringComparison]::Ordinal)
    if ($start -lt 0) {
        throw "${ScriptPath} does not define ${FunctionName}."
    }

    [int]$openBrace = $source.IndexOf("{", $start, [System.StringComparison]::Ordinal)
    [int]$depth = 0
    for ([int]$index = $openBrace; $index -lt $source.Length; $index++) {
        if ($source[$index] -eq "{") {
            $depth++
        }
        elseif ($source[$index] -eq "}") {
            $depth--
            if ($depth -eq 0) {
                return $source.Substring($start, $index - $start + 1)
            }
        }
    }

    throw "${ScriptPath} has an unclosed ${FunctionName} definition."
}

function Assert-Equal {
    param(
        [string]$Actual,
        [string]$Expected,
        [string]$Context
    )

    if ($Actual -ne $Expected) {
        throw "${Context}: expected [$Expected], actual [$Actual]."
    }
}

function Invoke-WindowsPowerShellNativeArgumentRoundTrip {
    param(
        [string]$Argument
    )

    [string]$childScriptPath = Join-Path ([System.IO.Path]::GetTempPath()) ("tloop-native-argument-round-trip-" + [Guid]::NewGuid().ToString("N") + ".ps1")
    Set-Content -LiteralPath $childScriptPath -Value '[Console]::Out.Write([Convert]::ToBase64String([Text.Encoding]::Unicode.GetBytes($args[0])))' -Encoding UTF8
    [string[]]$childArguments = @("-NoProfile", "-File", $childScriptPath, $Argument)
    [string[]]$escapedArguments = @($childArguments | ForEach-Object {
        ConvertTo-WindowsPowerShellNativeArgument -Argument $_
    })
    [string]$encodedOutput = & powershell.exe @escapedArguments
    if ($LASTEXITCODE -ne 0) {
        Remove-Item -LiteralPath $childScriptPath
        throw "Windows PowerShell native argument round trip failed."
    }

    Remove-Item -LiteralPath $childScriptPath
    return [Text.Encoding]::Unicode.GetString([Convert]::FromBase64String($encodedOutput))
}

[string[]]$definitions = @($ScriptPaths | ForEach-Object {
    Get-FunctionDefinition -ScriptPath $_ -FunctionName "ConvertTo-WindowsPowerShellNativeArgument"
})

foreach ($definition in @($definitions | Select-Object -Skip 1)) {
    Assert-Equal -Actual $definition -Expected $definitions[0] -Context "All E2E scripts must use the same native argument escaping implementation"
}

[string]$temporaryFunctionPath = Join-Path ([System.IO.Path]::GetTempPath()) "tloop-native-argument-escaping.ps1"
Set-Content -LiteralPath $temporaryFunctionPath -Value $definitions[0] -Encoding UTF8
. $temporaryFunctionPath
Remove-Item -LiteralPath $temporaryFunctionPath

Assert-Equal -Actual (ConvertTo-WindowsPowerShellNativeArgument -Argument '\"') -Expected '\\\"' -Context "Existing escaped quote"
Assert-Equal -Actual (ConvertTo-WindowsPowerShellNativeArgument -Argument 'C:\Users\Example Name\') -Expected 'C:\Users\Example Name\\' -Context "Trailing backslash in quoted argument"
Assert-Equal -Actual (ConvertTo-WindowsPowerShellNativeArgument -Argument '\\\"') -Expected '\\\\\\\"' -Context "Consecutive backslashes before quote"
Assert-Equal -Actual (ConvertTo-WindowsPowerShellNativeArgument -Argument 'C:\Users\Example\Project') -Expected 'C:\Users\Example\Project' -Context "Ordinary Windows path"
# Built from code points so this file stays ASCII: Windows PowerShell 5.1 decodes a script without a
# BOM in the system code page, and on a CJK code page such as 936 the UTF-8 bytes of a Japanese
# literal pair up with its closing quote, leaving the string unterminated.
[string]$japaneseDirectory = -join ([char[]](0x65E5, 0x672C, 0x8A9E))
[string]$japanesePath = -join ([char[]](0x30D1, 0x30B9))
[string]$japaneseValue = [string][char]0x5024
Assert-Equal -Actual (ConvertTo-WindowsPowerShellNativeArgument -Argument ($japaneseDirectory + '\' + $japanesePath + '"' + $japaneseValue)) -Expected ($japaneseDirectory + '\' + $japanesePath + '\"' + $japaneseValue) -Context "Japanese text"
Assert-Equal -Actual (ConvertTo-WindowsPowerShellNativeArgument -Argument 'C:\Users\Example\') -Expected 'C:\Users\Example\' -Context "Unquoted path with trailing backslash"
Assert-Equal -Actual (ConvertTo-WindowsPowerShellNativeArgument -Argument '\') -Expected '\' -Context "Unquoted standalone backslash"
Assert-Equal -Actual (ConvertTo-WindowsPowerShellNativeArgument -Argument "C:\Users\Example\`n") -Expected "C:\Users\Example\`n" -Context "Trailing backslash before line feed"
if ($PSVersionTable.PSVersion.Major -lt 6) {
    Assert-Equal -Actual (Invoke-WindowsPowerShellNativeArgumentRoundTrip -Argument 'C:\Users\Example\') -Expected 'C:\Users\Example\' -Context "Native process receives unquoted path with trailing backslash"
    Assert-Equal -Actual (Invoke-WindowsPowerShellNativeArgumentRoundTrip -Argument 'C:\Users\Example Name\') -Expected 'C:\Users\Example Name\' -Context "Native process receives quoted path with trailing backslash"
}

Write-Host "PowerShell native argument escaping tests passed."
