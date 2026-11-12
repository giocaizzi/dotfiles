[Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)

$os = Get-CimInstance Win32_OperatingSystem
$usedMemory = $os.TotalVisibleMemorySize - $os.FreePhysicalMemory
$memoryPercent = [math]::Round($usedMemory * 100 / $os.TotalVisibleMemorySize)
$memoryColor = if ($memoryPercent -ge 80) { "colour3" } else { "default" }

Write-Output (' #[fg=colour8][#[fg={0}]󰍛 {1}%#[fg=colour8]]' -f `
    $memoryColor, `
    $memoryPercent)

try {
    $disk = Get-CimInstance Win32_LogicalDisk -Filter ("DeviceID='{0}'" -f $env:SystemDrive) -ErrorAction SilentlyContinue |
        Select-Object -First 1
    if ($disk -and $disk.Size -gt 0 -and $null -ne $disk.FreeSpace) {
        $usedDisk = $disk.Size - $disk.FreeSpace
        $diskPercent = [math]::Round($usedDisk * 100 / $disk.Size)
        $diskColor = if ($diskPercent -ge 80) { "colour3" } else { "default" }

        Write-Output (' #[fg=colour8][#[fg={0}]󰋊 {1}%#[fg=colour8]]' -f `
            $diskColor, `
            $diskPercent)
    }
}
catch {
    # Disk status is optional; a query failure must not suppress other segments.
    $disk = $null
}

$cpuPercent = (Get-CimInstance Win32_Processor | Measure-Object -Property LoadPercentage -Average).Average
if ($null -ne $cpuPercent) {
    $cpuPercent = [math]::Round($cpuPercent)
    $cpuColor = if ($cpuPercent -ge 80) { "colour3" } else { "default" }
    Write-Output (' #[fg=colour8][#[fg={0}]󰻠 {1}%#[fg=colour8]]' -f $cpuColor, $cpuPercent)
}

$battery = Get-CimInstance Win32_Battery -ErrorAction SilentlyContinue | Select-Object -First 1
if ($battery) {
    $batteryColor = if ($battery.EstimatedChargeRemaining -lt 20) { "colour1" } else { "default" }
    Write-Output (' #[fg=colour8][#[fg={0}]󰁹 {1}%#[fg=colour8]]' -f `
        $batteryColor, `
        $battery.EstimatedChargeRemaining)
}
