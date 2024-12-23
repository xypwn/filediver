function InstallDependencies {
    param ()
    pip install -r requirements.txt
}

function SetupEnvironment {
    param ()
    $pyPresent = Get-Command py.exe
    if($pyPresent) {
        $pythonVersion = [semver]((py.exe -3.11 --version) -replace ".*\s")
        $python = "py.exe"
        $pyArgs = "-3.11"
    } else {
        $pythonVersion = [semver]((python.exe --version) -replace ".*\s")
        $python = "python.exe"
        $pyArgs = ""
    }
    if($pythonVersion.Major -lt 3 -or $pythonVersion.Minor -ne 11) {
        throw [System.Data.InvalidConstraintException]::new([string]::Format("Python 3.11.* is required (found {0}.{1})",$pythonVersion.Major,$pythonVersion.Minor))
    }
    if(![System.IO.File]::Exists(".venv")) {
        Write-Output "Creating vitual environment..."
        &$python $pyArgs -m venv .venv
    }
    &".venv\Scripts\Activate.ps1"
    try {
        InstallDependencies
    }
    finally {
        deactivate
    }
}


try {
    $originalDirectory = $PWD.Path
    Set-Location $PSScriptRoot
    SetupEnvironment
} catch [System.Management.Automation.CommandNotFoundException] {
    if($PSItem.TargetObject -eq "py.exe" -or $PSItem.TargetObject -eq "python.exe") {
        Write-Error "Python installation not found, please install python 3.11.*"
    } else {
        Write-Error $PSItem
        Write-Error $PSItem.TargetObject
    }
} catch [System.Data.InvalidConstraintException] {
    Write-Error $PSItem
} finally {
    Set-Location $originalDirectory
}