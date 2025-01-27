module github.com/sarchlab/akkalat

require (
	github.com/tebeka/atexit v0.3.0
	github.com/tidwall/pretty v1.0.2 // indirect
	gitlab.com/akita/akita/v3 v3.0.0-alpha.18
	gitlab.com/akita/mem/v3 v3.0.0-alpha.8
	gitlab.com/akita/mgpusim/v3 v3.0.0-alpha.1
	gitlab.com/akita/noc/v3 v3.0.0-alpha.8
)

// Make sure mgpusim is on the latest commit of
// `290-command-processor-reuses-todma-sender-port` branch
replace gitlab.com/akita/mgpusim/v3 => ../mgpusim

replace gitlab.com/akita/noc/v3 => ../noc

go 1.16
