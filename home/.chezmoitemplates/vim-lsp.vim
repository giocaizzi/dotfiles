{{- $lsp := (includeTemplate "agents-config.toml" . | fromToml).lsp | default dict -}}
" Language servers for yegappan/lsp. Rendered from [lsp.servers.*] in
" .chezmoitemplates/agents-config.toml: edit the servers there, not here.
" Installed as plugin/lsp.vim, so it loads before pack/*/start; the plugin fires
" `User LspSetup` once when it enables itself (on VimEnter).
let s:servers = json_decode('{{ ternary (get $lsp "servers" | default dict) dict (get $lsp "enabled" | default false) | toJson | replace "'" "''" }}')

function! s:Setup() abort
  let l:list = []
  for [l:name, l:s] in sort(items(s:servers))
    if !executable(l:s.command[0])
      continue
    endif
    let l:path = exepath(l:s.command[0])
    let l:args = l:s.command[1:]
    " On Windows an npm package's command is a .cmd shim, which a process spawn
    " without a shell cannot start, so it runs through cmd /c.
    if has('win32') && l:path =~? '\.\(cmd\|bat\)$'
      let l:args = ['/c', l:path] + l:args
      let l:path = 'cmd'
    endif
    " Vim filetypes: the server's, else the distinct language IDs.
    let l:filetypes = get(l:s, 'filetypes', uniq(sort(values(l:s.extensions))))
    " Formatting comes only from the `format = true` server: it is the explicit
    " documentFormatting provider, the others have the feature off (lsp-features).
    call add(l:list, #{
          \ name: l:name,
          \ filetype: l:filetypes,
          \ path: l:path,
          \ args: l:args,
          \ features: #{documentFormatting: get(l:s, 'format', 0) ? v:true : v:false},
          \ })
  endfor
  if !empty(l:list)
    call LspAddServer(l:list)
  endif
endfunction

" Format on save, only when a server is ready for the buffer.
function! s:Format() abort
  if exists('*LspServerReady') && LspServerReady()
    silent LspFormat
  endif
endfunction

augroup VimLsp
  autocmd!
  autocmd User LspSetup call s:Setup()
  for s:s in values(s:servers)
    if get(s:s, 'format', 0)
      for s:ext in keys(s:s.extensions)
        execute 'autocmd BufWritePre *' . s:ext . ' call s:Format()'
      endfor
    endif
  endfor
augroup END
