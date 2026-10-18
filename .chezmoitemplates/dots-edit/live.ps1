# dots-edit live preview pane for Windows (docs/OS.md#editing-your-dotfiles), the twin of
# live.sh, run by dots-edit in a Windows Terminal split pane:
# live.ps1 <source file> <edited file> <snapshot dir> <chezmoi> <source dir>.
# On every save of the edited file it copies it (re-encrypted when the source file is an
# age file) into the private source copy <snapshot dir>/shadow, renders this machine's
# target state from that copy and shows the diff of each target that differs from
# <snapshot dir>/before/here.json. Paths into the copy read as the real source dir. It
# ends, closing the pane, when <snapshot dir>/stop appears or the snapshot dir is gone.
param([string]$File, [string]$Edited, [string]$Dir, [string]$Cz, [string]$Src)
$shadow = Join-Path $Dir shadow
$last = $null
function Get-State([string]$Path) {
	$h = New-Object System.Collections.Hashtable ([StringComparer]::Ordinal)
	(Get-Content -Raw $Path | ConvertFrom-Json).PSObject.Properties | ForEach-Object { $h[$_.Name] = $_.Value }
	$h
}
function Get-Text($Entry) {
	if ($null -eq $Entry) { '' } elseif ($Entry.contents) { $Entry.contents } elseif ($Entry.linkname) { $Entry.linkname } else { '' }
}
$before = Get-State (Join-Path $Dir 'before/here.json')
while ((Test-Path $Dir) -and -not (Test-Path (Join-Path $Dir stop))) {
	$sum = (Get-FileHash -ErrorAction SilentlyContinue $Edited).Hash
	if ($sum -and $sum -ne $last) {
		$last = $sum
		$target = Join-Path $shadow $File
		if ($File -like '*.age') { & $Cz encrypt --output $target $Edited } else { Copy-Item -Force $Edited $target }
		try { Clear-Host } catch { }
		"dots-edit live preview: $File on this machine (refreshes on save)"
		$json = Join-Path $Dir live.json
		$err = Join-Path $Dir live.err
		$p = Start-Process -FilePath $Cz -ArgumentList '--source', "`"$shadow`"", 'dump', '--exclude', 'dirs,externals', '--format', 'json' `
			-RedirectStandardOutput $json -RedirectStandardError $err -NoNewWindow -Wait -PassThru
		if ($p.ExitCode -ne 0) { Get-Content $err; Start-Sleep 1; continue }
		# The copy's own path, in the dump as written and JSON-escaped, reads as the source dir.
		$text = Get-Content -Raw $json
		foreach ($from in @($shadow, ($shadow -replace '\\', '/'))) {
			$to = if ($from.Contains('\')) { $Src } else { $Src -replace '\\', '/' }
			$text = $text.Replace(($from -replace '\\', '\\'), ($to -replace '\\', '\\'))
		}
		[IO.File]::WriteAllText($json, $text)
		$now = Get-State $json
		$keys = @(@($before.Keys) + @($now.Keys) | Sort-Object -Unique -CaseSensitive)
		foreach ($k in $keys) {
			$a = Get-Text $before[$k]
			$b = Get-Text $now[$k]
			if ($a -cne $b) {
				[IO.File]::WriteAllText((Join-Path $Dir a), $a)
				[IO.File]::WriteAllText((Join-Path $Dir b), $b)
				"--- $k"
				git diff --no-index --no-color -- (Join-Path $Dir a) (Join-Path $Dir b) | Select-Object -Skip 4
			}
		}
	}
	Start-Sleep 1
}
