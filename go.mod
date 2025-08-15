module github.com/lrstanley/vex

go 1.25.0

// replace (
// 	github.com/charmbracelet/bubbletea/v2 => ./bubbletea
// 	github.com/charmbracelet/ultraviolet => ./ultraviolet
// )

replace (
	// TODO: https://github.com/charmbracelet/bubbles/pull/823
	github.com/charmbracelet/bubbles/v2 => github.com/lrstanley/bubbles/v2 v2.0.0-beta.1.0.20250809091447-b4884c2f81fc
	// TODO: https://github.com/charmbracelet/x/pull/516
	github.com/charmbracelet/x/exp/teatest/v2 => github.com/lrstanley/x/exp/teatest/v2 v2.0.0-20250731050217-1e52275d474b
)

require (
	github.com/Code-Hex/go-generics-cache v1.5.1
	github.com/alecthomas/chroma/v2 v2.20.0
	github.com/alecthomas/kong v1.12.0
	github.com/atotto/clipboard v0.1.4
	github.com/charmbracelet/bubbles/v2 v2.0.0-beta.1
	github.com/charmbracelet/bubbletea/v2 v2.0.0-beta.4.0.20250813213544-5cc219db8892
	github.com/charmbracelet/colorprofile v0.3.2
	github.com/charmbracelet/lipgloss/v2 v2.0.0-beta.3.0.20250814164412-7c497c73cf36
	github.com/charmbracelet/x/ansi v0.10.1
	github.com/charmbracelet/x/exp/teatest/v2 v2.0.0-20250725211024-d60e1b0112b2
	github.com/gkampitakis/go-snaps v0.5.14
	github.com/goccy/go-yaml v1.18.0
	github.com/gofrs/uuid/v5 v5.3.2
	github.com/hashicorp/vault/api v1.20.0
	github.com/joho/godotenv v1.5.1
	github.com/lrstanley/bubbletint/chromatint/v2 v2.0.0-alpha.0
	github.com/lrstanley/bubbletint/v2 v2.0.0-alpha.8
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/rivo/uniseg v0.4.7
	golang.org/x/sync v0.16.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/aymanbagabas/go-udiff v0.3.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/charmbracelet/ultraviolet v0.0.0-20250813213450-50737e162af5 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.14-0.20250505150409-97991a1f17d1 // indirect
	github.com/charmbracelet/x/exp/golden v0.0.0-20250806222409-83e3a29d542f // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/charmbracelet/x/termios v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.2.2 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/gkampitakis/ciinfo v0.3.3 // indirect
	github.com/gkampitakis/go-diff v1.3.2 // indirect
	github.com/go-jose/go-jose/v4 v4.1.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.2.0 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.7 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-7 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/maruel/natural v1.1.1 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/sahilm/fuzzy v0.1.1 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	golang.org/x/time v0.12.0 // indirect
)
