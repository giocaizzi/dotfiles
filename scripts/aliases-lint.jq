# Validates the alias registry (.chezmoidata/aliases.toml schema) dumped as JSON.
# Prints one error per line; no output means valid. Used by scripts/aliases-lint.sh.

def shells: ["posix", "powershell"];
def oses: ["darwin", "linux", "windows"];
def roles: ["personal", "work", "light"];
def keys_allowed: ["cmd", "posix", "powershell", "shells", "os", "machines", "why"];
def subset($allowed): all(.[]; . as $v | $allowed | index([$v]));

(.alias // {} | to_entries[] | .key as $n | .value as $a | ($a.shells // shells) as $on |
  (if ($n | test("^[A-Za-z0-9._][A-Za-z0-9._-]*$")) | not then "alias \($n): invalid name" else empty end),
  ($a | keys[] | select(. as $k | keys_allowed | index([$k]) | not) | "alias \($n): unknown key \(.)"),
  (if $a.cmd and ($a.posix or $a.powershell) then "alias \($n): use cmd or posix/powershell, not both" else empty end),
  (if ($on | type) != "array" or ($on | length) == 0 or ($on | subset(shells) | not) then "alias \($n): shells must be a non-empty subset of \(shells)" else empty end),
  (if $a.shells and ($a.shells | length) < 2 and ($a.why | not) then "alias \($n): limited to \($a.shells) without why" else empty end),
  ($on[]? | select(. as $s | $a.cmd or $a[$s] | not) | "alias \($n): no \(.) command (add one, or limit shells with why)"),
  (shells[] | select(. as $s | $a[$s] and ($on | index([$s]) | not)) | "alias \($n): \(.) command set but \(.) not in shells"),
  (if $a.os and ($a.os | subset(oses) | not) then "alias \($n): os must be a subset of \(oses)" else empty end),
  (if $a.machines and ($a.machines | subset(roles) | not) then "alias \($n): machines must be a subset of \(roles)" else empty end)
),
(.aliasBlock // {} | to_entries[] | select((.value | keys) != ["message"] or (.value.message | type) != "string") | "block \(.key): needs exactly a message string"),
(.aliasExternal // {} | to_entries[] | .key as $sh | .value | to_entries[] | select(.value | type != "boolean") | "external \($sh).\(.key): must be true or false")
