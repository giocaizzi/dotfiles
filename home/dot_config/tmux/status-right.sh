#!/bin/sh

memory=
battery=

case "$(uname -s)" in
  Darwin)
    page_size=$(sysctl -n hw.pagesize 2>/dev/null)
    total_bytes=$(sysctl -n hw.memsize 2>/dev/null)
    used_pages=$(vm_stat 2>/dev/null | awk '/Pages active|Pages wired down|Pages occupied by compressor/ { gsub(/\./, "", $NF); pages += $NF } END { print pages }')
    if [ -n "$page_size" ] && [ -n "$total_bytes" ] && [ -n "$used_pages" ]; then
      used_bytes=$((used_pages * page_size))
      memory="$(((used_bytes + 536870912) / 1073741824))/$(((total_bytes + 536870912) / 1073741824))GB:$((used_bytes * 100 / total_bytes))"
    fi
    battery=$(pmset -g batt 2>/dev/null | awk 'match($0, /[0-9]+%/) { print substr($0, RSTART, RLENGTH); exit }')
    ;;
  Linux)
    memory=$(free -b 2>/dev/null | awk '/^Mem:/ { used = $3; total = $2; printf "%d/%dGB:%d", (used + 536870912) / 1073741824, (total + 536870912) / 1073741824, used * 100 / total }')
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
    battery=$(powershell.exe -NoProfile -Command '$battery = Get-CimInstance Win32_Battery -ErrorAction SilentlyContinue | Select-Object -First 1; if ($battery) { "{0}%" -f $battery.EstimatedChargeRemaining }' 2>/dev/null)
    ;;
esac

if [ -n "$memory" ]; then
  memory_value=${memory%:*}
  memory_percent=${memory##*:}
  memory_color=#cdd6f4
  [ "$memory_percent" -ge 80 ] 2>/dev/null && memory_color=#f9e2af
  printf ' #[fg=#6c7086][#[fg=%s]󰍛 %s#[fg=#6c7086]]' "$memory_color" "$memory_value"
fi

if [ -n "$battery" ]; then
  battery_percent=${battery%%%}
  battery_color=#cdd6f4
  [ "$battery_percent" -lt 20 ] 2>/dev/null && battery_color=#f38ba8
  printf ' #[fg=#6c7086][#[fg=%s]󰁹 %s#[fg=#6c7086]]' "$battery_color" "$battery"
fi
