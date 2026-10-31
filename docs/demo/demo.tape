# Demo GIF for the README; regenerate with docs/demo/record.sh
Output demo.gif
Set Shell bash
Set FontSize 16
Set Width 960
Set Height 560
Set Padding 16
Set TypingSpeed 35ms
Set Theme "Dracula"

Hide
Type "export PATH=$PWD:$PATH PS1='$ ' && clear" Enter
Show

Type "cat config.toml" Enter
Sleep 4s
Type "clear" Enter
Type "cat template.eml" Enter
Sleep 2.5s
Type "clear" Enter
Type "gmt-mail -validate -config-path config.toml -template-path template.eml" Enter
Sleep 2s
Type "gmt-mail -dry-run -config-path config.toml -template-path template.eml | head -14" Enter
Sleep 4s
Type "clear" Enter
Type "gmt-mail -delay 500ms -config-path config.toml -template-path template.eml" Enter
Sleep 5s
