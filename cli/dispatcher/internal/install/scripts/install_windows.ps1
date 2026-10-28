$InstallDir = '{{INSTALL_DIR}}'
$ExpectedUloopPath = '{{EXPECTED_TLOOP_PATH}}'
$NormalizePath = {
    param([string]$Path)
    if (-not $Path) {
        return ''
    }
    return $Path.Trim().Trim('"').TrimEnd([char[]]@('\','/')).Replace('/','\')
}
function Set-UserPathWithInstallDirectoryFirst {
    param([string]$Directory)
    $UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $PathEntries = @()
    if ($UserPath) {
        $PathEntries = $UserPath -split ';' | Where-Object {
            $_ -and -not [string]::Equals((& $NormalizePath $_), (& $NormalizePath $Directory), [System.StringComparison]::OrdinalIgnoreCase)
        }
    }
    $NewUserPath = [string]::Join(';', @($Directory) + $PathEntries)
    if (-not [string]::Equals($UserPath, $NewUserPath, [System.StringComparison]::Ordinal)) {
        [Environment]::SetEnvironmentVariable('Path', $NewUserPath, 'User')
        Write-Host "Added $Directory to User PATH. Open a new terminal to use it everywhere."
    }
    $CurrentPathEntries = @()
    if ($env:Path) {
        $CurrentPathEntries = $env:Path -split ';' | Where-Object {
            $_ -and -not [string]::Equals((& $NormalizePath $_), (& $NormalizePath $Directory), [System.StringComparison]::OrdinalIgnoreCase)
        }
    }
    $env:Path = [string]::Join(';', @($Directory) + $CurrentPathEntries)
}
function Get-FirstUloopCommandFromPath {
    param([string]$PathValue)
    if (-not $PathValue) {
        return $null
    }
    foreach ($PathEntry in ($PathValue -split ';')) {
        $NormalizedPathEntry = & $NormalizePath $PathEntry
        if (-not $NormalizedPathEntry) {
            continue
        }
        if ($NormalizedPathEntry -match '^[A-Za-z]:$') {
            $NormalizedPathEntry = $NormalizedPathEntry + '\'
        }
        foreach ($ShimName in @('tloop.exe', 'tloop.cmd', 'tloop.ps1', 'tloop')) {
            $CandidatePath = Join-Path $NormalizedPathEntry $ShimName
            if (Test-Path $CandidatePath -PathType Leaf) {
                return $CandidatePath
            }
        }
    }
    return $null
}
function Report-PathShadowing {
    $MachinePath = [Environment]::GetEnvironmentVariable('Path', 'Machine')
    $UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $ResolvedPath = Get-FirstUloopCommandFromPath -PathValue ([string]::Join(';', @($MachinePath, $UserPath)))
    if (-not $ResolvedPath) {
        return
    }
    if ([string]::Equals($ResolvedPath, $ExpectedUloopPath, [System.StringComparison]::OrdinalIgnoreCase)) {
        return
    }
    Write-Host "Installed tloop to $ExpectedUloopPath, but PATH resolves tloop to:"
    Write-Host "  $ResolvedPath"
    Write-Host "Move $InstallDir earlier in PATH, or remove the legacy installation if it owns that command."
}
Set-UserPathWithInstallDirectoryFirst -Directory $InstallDir
Report-PathShadowing
