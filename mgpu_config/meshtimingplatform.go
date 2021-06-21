package runner

import (
	"fmt"

	"gitlab.com/akita/akita/v2/monitoring"
	"gitlab.com/akita/akita/v2/sim"
	"gitlab.com/akita/mem/v2/mem"
	"gitlab.com/akita/mem/v2/vm/mmu"
	"gitlab.com/akita/mgpusim/v2/driver"
	"gitlab.com/akita/noc/v2/networking/mesh"
)

// R9NanoPlatformBuilder can build a platform that equips R9Nano GPU.
type R9NanoPlatformBuilder struct {
	useParallelEngine bool
	debugISA          bool
	traceVis          bool
	visTraceStartTime sim.VTimeInSec
	visTraceEndTime   sim.VTimeInSec
	traceMem          bool
	numGPU            int
	log2PageSize      uint64
	monitor           *monitoring.Monitor
	meshSize          [3]int

	gpus []*GPU
}

// MakeR9NanoBuilder creates a EmuBuilder with default parameters.
func MakeR9NanoBuilder() R9NanoPlatformBuilder {
	b := R9NanoPlatformBuilder{
		numGPU:            4,
		log2PageSize:      12,
		visTraceStartTime: -1,
		visTraceEndTime:   -1,
		meshSize:          [3]int{1, 1, 1},
	}

	// Minus one since CPU is at (0,0,0)
	//b.numGPU = (b.meshSize[0] * b.meshSize[1] * b.meshSize[2]) - 1

	return b
}

/*
// WithParallelEngine lets the EmuBuilder to use parallel engine.
func (b R9NanoPlatformBuilder) WithParallelEngine() R9NanoPlatformBuilder {
	b.useParallelEngine = true
	return b
}

// WithISADebugging enables ISA debugging in the simulation.
func (b R9NanoPlatformBuilder) WithISADebugging() R9NanoPlatformBuilder {
	b.debugISA = true
	return b
}

// WithVisTracing lets the platform to record traces for visualization purposes.
func (b R9NanoPlatformBuilder) WithVisTracing() R9NanoPlatformBuilder {
	b.traceVis = true
	return b
}

// WithPartialVisTracing lets the platform to record traces for visualization
// purposes. The trace will only be collected from the start time to the end
// time.
func (b R9NanoPlatformBuilder) WithPartialVisTracing(
	start, end sim.VTimeInSec,
) R9NanoPlatformBuilder {
	b.traceVis = true
	b.visTraceStartTime = start
	b.visTraceEndTime = end

	return b
}

// WithMemTracing lets the platform to trace memory operations.
func (b R9NanoPlatformBuilder) WithMemTracing() R9NanoPlatformBuilder {
	b.traceMem = true
	return b
}

// WithNumGPU sets the number of GPUs to build.
func (b R9NanoPlatformBuilder) WithNumGPU(n int) R9NanoPlatformBuilder {
	b.numGPU = n
	return b
}

// WithLog2PageSize sets the page size as a power of 2.
func (b R9NanoPlatformBuilder) WithLog2PageSize(
	n uint64,
) R9NanoPlatformBuilder {
	b.log2PageSize = n
	return b
}

// WithMonitor sets the monitor that is used to monitor the simulation
func (b R9NanoPlatformBuilder) WithMonitor(
	m *monitoring.Monitor,
) R9NanoPlatformBuilder {
	b.monitor = m
	return b
}
*/

// WithMeshWidth sets the width of GPUs in mesh.
func (b R9NanoPlatformBuilder) WithMesh(mesh bool, n [3]int) R9NanoPlatformBuilder {

	if mesh {
		b.meshSize = n
		b.numGPU = (b.meshSize[0] * b.meshSize[1] * b.meshSize[2]) - 1
	}

	//fmt.Println("WithMesh: Mesh size ", b.meshSize)
	//fmt.Println("WithMesh: Num GPUs ", b.numGPU)

	return b
}

// meshBuild builds a platform with R9Nano GPUs in a mesh
func (b R9NanoPlatformBuilder) meshBuild() *Platform {
	engine := b.createEngine()
	if b.monitor != nil {
		b.monitor.RegisterEngine(engine)
	}

	mmuComponent, pageTable := b.createMMU(engine)

	gpuDriver := driver.NewDriver(engine, pageTable, b.log2PageSize)
	// file, err := os.Create("driver_comm.csv")
	// if err != nil {
	// 	panic(err)
	// }
	// gpuDriver.GetPortByName("GPU").AcceptHook(
	// 	sim.NewPortMsgLogger(log.New(file, "", 0)))

	if b.monitor != nil {
		b.monitor.RegisterComponent(gpuDriver)
	}

	gpuBuilder := b.createMeshGPUBuilder(engine, gpuDriver, mmuComponent)

	meshConnector := b.createMeshConnection(engine, gpuDriver, mmuComponent)

	mmuComponent.MigrationServiceProvider = gpuDriver.GetPortByName("MMU")

	rdmaAddressTable := b.createRDMAAddrTable()
	pmcAddressTable := b.createPMCPageTable()

	b.createMeshGPUs(meshConnector, gpuBuilder, gpuDriver, rdmaAddressTable,
		pmcAddressTable)

	meshConnector.EstablishNetwork()

	return &Platform{
		Engine: engine,
		Driver: gpuDriver,
		GPUs:   b.gpus,
	}
}

func (b *R9NanoPlatformBuilder) createMeshGPUs(
	meshConnector *mesh.Connector,
	gpuBuilder R9NanoGPUBuilder,
	gpuDriver *driver.Driver,
	rdmaAddressTable *mem.BankedLowModuleFinder,
	pmcAddressTable *mem.BankedLowModuleFinder,
) {
	for z := 0; z < b.meshSize[2]; z++ {
		for y := 0; y < b.meshSize[1]; y++ {
			for x := 0; x < b.meshSize[0]; x++ {

				// Reserving (0,0,0) for the CPU
				if x != 0 || y != 0 || z != 0 {

					i := x + (y * b.meshSize[0]) + (z * b.meshSize[0] * b.meshSize[1])

					element := [3]int{x, y, z}
					b.createMeshGPU(i, gpuBuilder, gpuDriver,
						rdmaAddressTable, pmcAddressTable,
						meshConnector, element)

					//fmt.Println("createMeshGPUs: GPU ", i, element, " GPUs", len(b.gpus))
					//meshConnector.AddTile([3]int{x, y, z}, b.gpus[i].Domain.Ports())
					//fmt.Println("createMeshGPUs: GPU ", i, " Ports", b.gpus[i-1].Domain.Ports())
				}
			}
		}
	}
}

/*
func (b R9NanoPlatformBuilder) createPMCPageTable() *mem.BankedLowModuleFinder {
	pmcAddressTable := new(mem.BankedLowModuleFinder)
	pmcAddressTable.BankSize = 4 * mem.GB
	pmcAddressTable.LowModules = append(pmcAddressTable.LowModules, nil)
	return pmcAddressTable
}

func (b R9NanoPlatformBuilder) createRDMAAddrTable() *mem.BankedLowModuleFinder {
	rdmaAddressTable := new(mem.BankedLowModuleFinder)
	rdmaAddressTable.BankSize = 4 * mem.GB
	rdmaAddressTable.LowModules = append(rdmaAddressTable.LowModules, nil)
	return rdmaAddressTable
}
*/

func (b R9NanoPlatformBuilder) createMeshConnection(
	engine sim.Engine,
	gpuDriver *driver.Driver,
	mmuComponent *mmu.MMUImpl,
) *mesh.Connector {
	//connection := sim.NewDirectConnection(engine)
	// connection := noc.NewFixedBandwidthConnection(32, engine, 1*sim.GHz)
	// connection.SrcBufferCapacity = 40960000
	meshConnector := mesh.NewConnector().
		WithEngine(engine).
		WithSwitchLatency(20)
	meshConnector.CreateNetwork("Mesh")

	meshConnector.AddTile([3]int{0, 0, 0},
		[]sim.Port{
			gpuDriver.GetPortByName("GPU"),
			gpuDriver.GetPortByName("MMU"),
			mmuComponent.GetPortByName("Migration"),
			mmuComponent.GetPortByName("Top"),
		})
	/*
		rootComplexID := meshConnector.AddRootComplex(
			[]sim.Port{
				gpuDriver.GetPortByName("GPU"),
				gpuDriver.GetPortByName("MMU"),
				mmuComponent.GetPortByName("Migration"),
				mmuComponent.GetPortByName("Top"),
			})
	*/
	return meshConnector
}

/*
func (b R9NanoPlatformBuilder) createEngine() sim.Engine {
	var engine sim.Engine

	if b.useParallelEngine {
		engine = sim.NewParallelEngine()
	} else {
		engine = sim.NewSerialEngine()
	}
	// engine.AcceptHook(sim.NewEventLogger(log.New(os.Stdout, "", 0)))

	return engine
}

func (b R9NanoPlatformBuilder) createMMU(
	engine sim.Engine,
) (*mmu.MMUImpl, vm.PageTable) {
	pageTable := vm.NewPageTable(b.log2PageSize)
	mmuBuilder := mmu.MakeBuilder().
		WithEngine(engine).
		WithFreq(1 * sim.GHz).
		WithPageWalkingLatency(100).
		WithLog2PageSize(b.log2PageSize).
		WithPageTable(pageTable)

	mmuComponent := mmuBuilder.Build("MMU")

	if b.monitor != nil {
		b.monitor.RegisterComponent(mmuComponent)
	}

	return mmuComponent, pageTable
}
*/

func (b *R9NanoPlatformBuilder) createMeshGPUBuilder(
	engine sim.Engine,
	gpuDriver *driver.Driver,
	mmuComponent *mmu.MMUImpl,
) R9NanoGPUBuilder {
	gpuBuilder := MakeR9NanoGPUBuilder().
		WithEngine(engine).
		WithMMU(mmuComponent).
		WithNumCUPerShaderArray(4).
		WithNumShaderArray(1).
		WithNumMemoryBank(1).
		WithLog2MemoryBankInterleavingSize(7).
		WithLog2PageSize(b.log2PageSize)

	if b.monitor != nil {
		gpuBuilder = gpuBuilder.WithMonitor(b.monitor)
	}

	gpuBuilder = b.setVisTracer(gpuDriver, gpuBuilder)
	gpuBuilder = b.setMemTracer(gpuBuilder)
	gpuBuilder = b.setISADebugger(gpuBuilder)

	return gpuBuilder
}

/*
func (b *R9NanoPlatformBuilder) setISADebugger(
	gpuBuilder R9NanoGPUBuilder,
) R9NanoGPUBuilder {
	if !b.debugISA {
		return gpuBuilder
	}

	gpuBuilder = gpuBuilder.WithISADebugging()
	return gpuBuilder
}

func (b *R9NanoPlatformBuilder) setMemTracer(
	gpuBuilder R9NanoGPUBuilder,
) R9NanoGPUBuilder {
	if !b.traceMem {
		return gpuBuilder
	}

	file, err := os.Create("mem.trace")
	if err != nil {
		panic(err)
	}
	logger := log.New(file, "", 0)
	memTracer := memtraces.NewTracer(logger)
	gpuBuilder = gpuBuilder.WithMemTracer(memTracer)
	return gpuBuilder
}

func (b *R9NanoPlatformBuilder) setVisTracer(
	gpuDriver *driver.Driver,
	gpuBuilder R9NanoGPUBuilder,
) R9NanoGPUBuilder {
	if !b.traceVis {
		return gpuBuilder
	}

	tracer := tracing.NewMySQLTracerWithTimeRange(
		b.visTraceStartTime,
		b.visTraceEndTime)
	tracer.Init()
	tracing.CollectTrace(gpuDriver, tracer)

	gpuBuilder = gpuBuilder.WithVisTracer(tracer)
	return gpuBuilder
}
*/

func (b *R9NanoPlatformBuilder) createMeshGPU(
	index int,
	gpuBuilder R9NanoGPUBuilder,
	gpuDriver *driver.Driver,
	rdmaAddressTable *mem.BankedLowModuleFinder,
	pmcAddressTable *mem.BankedLowModuleFinder,
	meshConnector *mesh.Connector,
	element [3]int,
) *GPU {
	name := fmt.Sprintf("GPU%d", index)
	memAddrOffset := uint64(index) * 4 * mem.GB
	gpu := gpuBuilder.
		WithMemAddrOffset(memAddrOffset).
		Build(name, uint64(index))

	//fmt.Println("createMeshGPU: Mem Addr Offset ", index, memAddrOffset)
	gpuDriver.RegisterGPU(gpu.Domain.GetPortByName("CommandProcessor"),
		4*mem.GB)
	gpu.CommandProcessor.Driver = gpuDriver.GetPortByName("GPU")

	b.configRDMAEngine(gpu, rdmaAddressTable)
	b.configPMC(gpu, gpuDriver, pmcAddressTable)

	meshConnector.AddTile(element, gpu.Domain.Ports())

	b.gpus = append(b.gpus, gpu)

	return gpu
}

/*
func (b *R9NanoPlatformBuilder) configRDMAEngine(
	gpu *GPU,
	addrTable *mem.BankedLowModuleFinder,
) {
	gpu.RDMAEngine.RemoteRDMAAddressTable = addrTable
	addrTable.LowModules = append(
		addrTable.LowModules,
		gpu.RDMAEngine.ToOutside)
}

func (b *R9NanoPlatformBuilder) configPMC(
	gpu *GPU,
	gpuDriver *driver.Driver,
	addrTable *mem.BankedLowModuleFinder,
) {
	gpu.PMC.RemotePMCAddressTable = addrTable
	addrTable.LowModules = append(
		addrTable.LowModules,
		gpu.PMC.GetPortByName("Remote"))
	gpuDriver.RemotePMCPorts = append(
		gpuDriver.RemotePMCPorts, gpu.PMC.GetPortByName("Remote"))
}
*/
