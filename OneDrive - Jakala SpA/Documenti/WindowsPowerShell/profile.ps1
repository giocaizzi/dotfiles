# conda
#region conda initialize
# !! Contents within this block are managed by 'conda init' !!
If (Test-Path "C:\Users\gcaizzi\AppData\Local\miniconda3\Scripts\conda.exe") {
    (& "C:\Users\gcaizzi\AppData\Local\miniconda3\Scripts\conda.exe" "shell.powershell" "hook") | Out-String | ?{$_} | Invoke-Expression
}
#endregion

# oh-my-posh settings
oh-my-posh init pwsh --config "~\.mytheme.omp.json" | Invoke-Expression
