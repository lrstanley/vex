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
Set Framerate 15
Set TypingSpeed 120ms
Env PS1 "$ "

Type "./vex"
Sleep 500ms
Enter
Sleep 1s

# help dialog
Type "?"
Sleep 1s
Type "?"

# filter for "kv-v1-1"
Type "/"
Sleep 500ms
Type "kv-v1-1"
Enter
Sleep 500ms
Type "d" # open details dialog
Sleep 1s
Type "d" # close details dialog
Sleep 500ms
Enter # select mount

# filter for "this/"
Sleep 500ms
Type "/"
Sleep 250ms
Type "this"
Enter # dive
Sleep 500ms
Enter # dive
Sleep 500ms
Enter # dive
Sleep 500ms
Enter # dive
Sleep 500ms

# switch to secrets view
Type ":"
Sleep 500ms
Type "secr"
Sleep 500ms
Enter

# filter for "this/"
Sleep 500ms
Type "/"
Sleep 250ms
Type "this"
Enter # dive
Sleep 500ms

# switch to aclpolicies view
Type ":"
Sleep 500ms
Type "acl"
Sleep 500ms
Enter # switch to aclpolicies view

# filter for "default"
Type "/"
Sleep 500ms
Type "default"
Enter
Sleep 500ms
Enter
Down@50ms 15
Sleep 500ms

# demo themes
Type ":"
Type@1ms "[[["
Sleep 1s
Type@1s "]]]]]"

Sleep 1s
