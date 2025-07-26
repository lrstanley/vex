# "task tape" or: go run github.com/charmbracelet/vhs@latest demo.tape
Output tmp/demo.gif
Require go
Require task

Set Theme "Dracula+"
Set FontSize 27
Set Width 1862
Set Height 1064
Set Margin 35
Set Padding 20
Set MarginFill "#674EFF"
Set BorderRadius 10
Set Framerate 20
Set TypingSpeed 200ms
Env PS1 "$ "

Type "./vex"
Sleep 500ms
Enter
Sleep 2s

# help dialog
Type "?"
Sleep 2s
Type "?"
Sleep 1s

# filter for "kv-v1-1"
Type "/"
Sleep 500ms
Type "kv-v1-1"
Enter
Sleep 1s
Enter # select mount

# filter for "this/"
Sleep 1s
Type "/"
Sleep 500ms
Type "this"
Enter # dive
Sleep 1s
Enter # dive
Sleep 1s
Enter # dive
Sleep 1s
Enter # dive
Sleep 1s

# switch to aclpolicies view
Type ":"
Sleep 500ms
Type "acl"
Sleep 1s
Enter # switch to aclpolicies view

# filter for "default"
Type "/"
Sleep 500ms
Type "default"
Enter
Sleep 1s
Enter
Down@50ms 30
Sleep 2s

# demo themes
Type ":"
Sleep 500ms
Type@1ms "[[["
Sleep 1s
Type@2s "]]]]]"

Sleep 2s
