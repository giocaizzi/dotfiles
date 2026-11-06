" Language servers for yegappan/lsp. Rendered from [harnessPolicy.lsp.server.*]
" (.chezmoidata/harness-policy.toml): edit the servers there, not here.
" Installed as plugin/lsp.vim, so it loads before pack/*/start; the plugin fires
" `User LspSetup` once when it enables itself (on VimEnter).
let s:servers = json_decode('{{ includeTemplate "lsp-servers" . | replace "'" "''" }}')

function! s:Setup() abort
  let l:list = []
  for [l:name, l:s] in sort(items(s:servers))
    if !executable(l:s.command[0])
      continue
    endif
    call add(l:list, #{
          \ name: l:name,
          \ filetype: uniq(sort(values(l:s.extensions))),
          \ path: exepath(l:s.command[0]),
          \ args: l:s.command[1:],
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
