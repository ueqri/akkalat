module github.com/sarchlab/akkalat

require (
	github.com/tebeka/atexit v0.3.0
	gitlab.com/akita/akita/v3 v3.0.0-alpha.11
	gitlab.com/akita/mem/v3 v3.0.0-alpha.1
	gitlab.com/akita/mgpusim/v3 v3.0.0-alpha.1
	gitlab.com/akita/noc/v3 v3.0.0-alpha.3
)

// Make sure mgpusim is on the latest commit of
// `290-command-processor-reuses-todma-sender-port` branch
replace gitlab.com/akita/mgpusim/v3 => ../mgpusim

// Make sure noc is on the latest commit of
// `12-implement-robust-meshnetworktracer` branch
replace gitlab.com/akita/noc/v3 => ../noc

replace gitlab.com/akita/akita/v3 => ../akita

go 1.16
