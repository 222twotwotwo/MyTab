param(
  [Parameter(Mandatory = $true)]
  [ValidateSet("running", "free")]
  [string]$Mode,

  [int]$StartPort = 8080,
  [int]$EndPort = 8099,
  [string]$PortFile = "",
  [switch]$All
)

function Test-MyTabServer {
  param([int]$Port)

  try {
    $response = Invoke-WebRequest -UseBasicParsing -TimeoutSec 1 -Uri "http://127.0.0.1:$Port/healthz"
    if ($response.StatusCode -ne 200) {
      return $false
    }

    $marker = $response.Headers["X-MyTab"]
    if ($marker -eq "true" -or $marker -contains "true") {
      return $true
    }

    return $response.Content -match '"status"\s*:\s*"ok"'
  } catch {
    return $false
  }
}

function Test-CanListen {
  param([int]$Port)

  $listener = $null
  try {
    $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, $Port)
    $listener.Start()
    return $true
  } catch {
    return $false
  } finally {
    if ($null -ne $listener) {
      $listener.Stop()
    }
  }
}

function Get-CachedPorts {
  if ([string]::IsNullOrWhiteSpace($PortFile) -or -not (Test-Path -LiteralPath $PortFile)) {
    return @()
  }

  Get-Content -LiteralPath $PortFile -ErrorAction SilentlyContinue |
    ForEach-Object {
      $value = 0
      if ([int]::TryParse($_.Trim(), [ref]$value) -and $value -gt 0) {
        $value
      }
    }
}

function Get-EphemeralPort {
  $listener = $null
  try {
    $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, 0)
    $listener.Start()
    return $listener.LocalEndpoint.Port
  } finally {
    if ($null -ne $listener) {
      $listener.Stop()
    }
  }
}

$preferredPorts = $StartPort..$EndPort

if ($Mode -eq "running") {
  $candidatePorts = @(Get-CachedPorts)
  $candidatePorts += $StartPort
  $candidatePorts = $candidatePorts | Sort-Object -Unique

  $found = @()
  foreach ($port in $candidatePorts) {
    if (Test-MyTabServer -Port $port) {
      if (-not $All) {
        Write-Output $port
        exit 0
      }
      $found += $port
    }
  }

  if ($found.Count -gt 0) {
    $found | ForEach-Object { Write-Output $_ }
    exit 0
  }

  exit 1
}

foreach ($port in $preferredPorts) {
  if (Test-CanListen -Port $port) {
    Write-Output $port
    exit 0
  }
}

$port = Get-EphemeralPort
if ($port -gt 0) {
  Write-Output $port
  exit 0
}

exit 1
