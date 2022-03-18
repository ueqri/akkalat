module github.com/sarchlab/akkalat

require (
	github.com/tebeka/atexit v0.3.0
	gitlab.com/akita/akita/v3 v3.0.0-alpha.9
	gitlab.com/akita/mem/v3 v3.0.0-alpha.1
	gitlab.com/akita/mgpusim/v3 v3.0.0-alpha.1
	gitlab.com/akita/noc/v3 v3.0.0-alpha.1
)

// Make sure mgpusim is on the latest commit of
// `90-command-processor-reuses-todma-sender-port` branch
replace gitlab.com/akita/mgpusim/v3 => ../mgpusim

// Make sure noc is on the latest commit of `9-task-tracing-for-networks` branch
replace gitlab.com/akita/noc/v3 => ../noc

// Make sure util is on the latest commit of `v2` branch
// replace gitlab.com/akita/util/v2 => ../util

go 1.16
