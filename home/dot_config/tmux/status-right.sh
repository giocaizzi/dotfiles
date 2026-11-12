#!/bin/sh

memory=
cpu=
battery=
disk=

case "$(uname -s)" in
  Darwin)
    page_size=$(sysctl -n hw.pagesize 2>/dev/null)
    total_bytes=$(sysctl -n hw.memsize 2>/dev/null)
    used_pages=$(vm_stat 2>/dev/null | awk '/Pages active|Pages wired down|Pages occupied by compressor/ { gsub(/\./, "", $NF); pages += $NF } END { print pages }')
    if [ -n "$page_size" ] && [ -n "$total_bytes" ] && [ -n "$used_pages" ]; then
      used_bytes=$((used_pages * page_size))
      memory="$(((used_bytes + 536870912) / 1073741824))/$(((total_bytes + 536870912) / 1073741824))GB:$((used_bytes * 100 / total_bytes))"
    fi
    cpu=$(top -l 1 -n 0 2>/dev/null | awk '/^CPU usage:/ { user = $3; sub(/%/, "", user); system = $5; sub(/%/, "", system); printf "%d", user + system; exit }')
    battery=$(pmset -g batt 2>/dev/null | awk 'match($0, /[0-9]+%/) { print substr($0, RSTART, RLENGTH); exit }')
    disk=$(df -Pk / 2>/dev/null | awk 'NR == 2 && $2 ~ /^[0-9]+$/ && $3 ~ /^[0-9]+$/ && $2 > 0 { used = $3 * 1024; total = $2 * 1024; printf "%d/%dGB:%d", (used + 536870912) / 1073741824, (total + 536870912) / 1073741824, used * 100 / total }')
    ;;
  Linux)
    memory=$(free -b 2>/dev/null | awk '/^Mem:/ { used = $3; total = $2; printf "%d/%dGB:%d", (used + 536870912) / 1073741824, (total + 536870912) / 1073741824, used * 100 / total }')
    cpu=$(LC_ALL=C top -bn1 2>/dev/null | awk '/^%Cpu\(s\):/ { idle = $8; sub(/,/, "", idle); printf "%d", 100 - idle; exit }')
    disk=$(df -Pk / 2>/dev/null | awk 'NR == 2 && $2 ~ /^[0-9]+$/ && $3 ~ /^[0-9]+$/ && $2 > 0 { used = $3 * 1024; total = $2 * 1024; printf "%d/%dGB:%d", (used + 536870912) / 1073741824, (total + 536870912) / 1073741824, used * 100 / total }')
    for capacity in /sys/class/power_supply/BAT*/capacity; do
      if [ -r "$capacity" ]; then
        battery="$(cat "$capacity")%"
        break
      fi
    done
    ;;
  MINGW*|MSYS*|CYGWIN*)
    # shellcheck disable=SC2016 # PowerShell variables must remain literal.
    memory=$(powershell.exe -NoProfile -Command '$os = Get-CimInstance Win32_OperatingSystem; $used = $os.TotalVisibleMemorySize - $os.FreePhysicalMemory; "{0}/{1}GB:{2}" -f [math]::Round($used / 1MB), [math]::Round($os.TotalVisibleMemorySize / 1MB), [math]::Round($used * 100 / $os.TotalVisibleMemorySize)' 2>/dev/null)
    # shellcheck disable=SC2016 # PowerShell variables must remain literal.
    cpu=$(powershell.exe -NoProfile -Command '(Get-CimInstance Win32_Processor | Measure-Object -Property LoadPercentage -Average).Average' 2>/dev/null)
    # shellcheck disable=SC2016 # PowerShell variables must remain literal.
    battery=$(powershell.exe -NoProfile -Command '$battery = Get-CimInstance Win32_Battery -ErrorAction SilentlyContinue | Select-Object -First 1; if ($battery) { "{0}%" -f $battery.EstimatedChargeRemaining }' 2>/dev/null)
    # shellcheck disable=SC2016 # PowerShell variables must remain literal.
    disk=$(powershell.exe -NoProfile -Command '$disk = Get-CimInstance Win32_LogicalDisk -Filter ("DeviceID={0}{1}{0}" -f [char]39, $env:SystemDrive) -ErrorAction SilentlyContinue | Select-Object -First 1; if ($disk -and $disk.Size -gt 0 -and $null -ne $disk.FreeSpace) { $used = $disk.Size - $disk.FreeSpace; "{0}/{1}GB:{2}" -f [math]::Round($used / 1GB), [math]::Round($disk.Size / 1GB), [math]::Round($used * 100 / $disk.Size) }' 2>/dev/null)
    ;;
esac

if [ -n "$memory" ]; then
  memory_percent=${memory##*:}
  memory_color=default
  [ "$memory_percent" -ge 80 ] 2>/dev/null && memory_color=colour3
  printf ' #[fg=colour8][#[fg=%s]󰍛 %s%%#[fg=colour8]]' "$memory_color" "$memory_percent"
fi

if [ -n "$disk" ]; then
  disk_percent=${disk##*:}
  disk_color=default
  [ "$disk_percent" -ge 80 ] 2>/dev/null && disk_color=colour3
  printf ' #[fg=colour8][#[fg=%s]󰋊 %s%%#[fg=colour8]]' "$disk_color" "$disk_percent"
fi

if [ -n "$cpu" ]; then
  cpu_color=default
  [ "$cpu" -ge 80 ] 2>/dev/null && cpu_color=colour3
  printf ' #[fg=colour8][#[fg=%s]󰻠 %s%%#[fg=colour8]]' "$cpu_color" "$cpu"
fi

if [ -n "$battery" ]; then
  battery_percent=${battery%%%}
  battery_color=default
  [ "$battery_percent" -lt 20 ] 2>/dev/null && battery_color=colour1
  printf ' #[fg=colour8][#[fg=%s]󰁹 %s#[fg=colour8]]' "$battery_color" "$battery"
fi
