$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $root

$serverUrl = 'http://localhost:34115'
$probeUrl = 'http://127.0.0.1:8765'

function Wait-ForServer {
    param(
        [string]$Url,
        [int]$TimeoutSeconds = 60
    )

    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        try {
            $response = Invoke-WebRequest -Uri $Url -Method Get -TimeoutSec 5 -UseBasicParsing -ErrorAction Stop
            if ($response.StatusCode -eq 200) {
                return $true
            }
        }
        catch {
            Start-Sleep -Seconds 1
        }
    }

    return $false
}

Write-Host 'Checking whether WIP is already running...'
$serverReady = $false
try {
    $response = Invoke-WebRequest -Uri "$serverUrl/api/apps" -Method Get -TimeoutSec 5 -UseBasicParsing -ErrorAction Stop
    if ($response.StatusCode -eq 200) {
        $serverReady = $true
    }
}
catch {
    $serverReady = $false
}

if (-not $serverReady) {
    Write-Host 'Starting WIP server...'
    $serverProcess = Start-Process -FilePath 'powershell' -ArgumentList @(
        '-NoProfile',
        '-ExecutionPolicy', 'Bypass',
        '-Command', "Set-Location '$root'; go run ."
    ) -WorkingDirectory $root -PassThru -WindowStyle Hidden

    if (-not (Wait-ForServer -Url "$serverUrl/api/apps" -TimeoutSeconds 90)) {
        throw 'WIP server did not become ready on http://localhost:34115 within the timeout.'
    }

    Write-Host "WIP server started with PID $($serverProcess.Id)"
}
else {
    Write-Host 'WIP is already running; reusing the existing server.'
}

$tempDir = Join-Path $env:TEMP ('wip-live-test-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

try {
    Write-Host 'Creating a temporary app...'
    $createBody = @{
        name        = 'LiveTest'
        description = 'Live process smoke test'
        localPath   = $tempDir
        createNew   = $true
    } | ConvertTo-Json -Compress

    $app = Invoke-RestMethod -Uri "$serverUrl/api/apps" -Method Post -ContentType 'application/json' -Body $createBody
    $appId = $app.id
    Write-Host "Created app ID: $appId"

    Write-Host 'Configuring a component to run a local Python server...'
    $componentBody = @{
        components = @(
            @{
                name         = 'App'
                startCommand = 'python -m http.server 8765 --bind 127.0.0.1'
                stopCommand  = ''
                runMode      = 'native'
            }
        )
    } | ConvertTo-Json -Compress

    Invoke-RestMethod -Uri "$serverUrl/api/apps/$appId/components" -Method Put -ContentType 'application/json' -Body $componentBody | Out-Null

    Write-Host 'Starting the app component...'
    $startBody = '{"component":"App"}'
    Invoke-RestMethod -Uri "$serverUrl/api/apps/$appId/start" -Method Post -ContentType 'application/json' -Body $startBody | Out-Null

    $deadline = (Get-Date).AddSeconds(20)
    $started = $false
    while ((Get-Date) -lt $deadline) {
        try {
            $probe = Invoke-WebRequest -Uri $probeUrl -Method Get -TimeoutSec 2 -UseBasicParsing -ErrorAction Stop
            if ($probe.StatusCode -eq 200) {
                $started = $true
                break
            }
        }
        catch {
            Start-Sleep -Seconds 1
        }
    }

    if (-not $started) {
        throw 'The started component did not begin responding on http://127.0.0.1:8765.'
    }

    Write-Host 'The app is responding on port 8765.'

    Write-Host 'Stopping the app component...'
    $stopBody = '{"component":"App"}'
    Invoke-RestMethod -Uri "$serverUrl/api/apps/$appId/stop" -Method Post -ContentType 'application/json' -Body $stopBody | Out-Null

    Start-Sleep -Seconds 2
    $portOpen = $false
    try {
        $check = Test-NetConnection -ComputerName '127.0.0.1' -Port 8765 -InformationLevel Quiet -WarningAction SilentlyContinue
        if ($check) {
            $portOpen = $true
        }
    }
    catch {
        $portOpen = $false
    }

    if ($portOpen) {
        throw 'Port 8765 is still open after the stop call.'
    }

    Write-Host 'Port 8765 is closed after stop.'
    Write-Host "Smoke test passed for app ID: $appId"
}
finally {
    Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue
}
