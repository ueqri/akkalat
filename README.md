# akkalat

Akkalat is the infrastructure to simulate wafer-scale GPUs.

## CU-level mesh with specific caches (v3)

After profiling the v2 model in [Daisen](https://osf.io/73ry8/), we found the **Instruction Caches** and **Scalar Caches** are definitely needed to avoid massive global memory requests. So we attached these specific caches inside each tile.

This is the **stable** version recently, we would do further research based on the visualization results of [Vis4Mesh](https://github.com/ueqri/vis4mesh). Further steps includes reorganizing DMA engine, redesigning memory address layout and so on.

### Build

```bash
git clone git@github.com:ueqri/akkalat.git
cd akkalat
git checkout add-caches-for-tile
cd samples/fir
go build

# Situation 1: run FIR without NoC tracing
./fir -length=100000 -verify -timing

# Situation 2: run FIR with tracing and Vis4Mesh
# customize the env to meet your demands
export AKITA_TRACE_PASSWORD=akitavis
export AKITA_TRACE_IP=127.0.0.1
export AKITA_TRACE_PORT=6379
export AKITA_TRACE_REDIS_DB=0
./fir -length=100000 -verify -timing -trace-vis
# then follow Vis4Mesh tutorial to run visualization
```

## CU-level mesh without any cache (v2)

### Synopsis

In this part, we propose a new **tile** model without any cache inside. We implement each **tile** with a CU, SRAM(Interleaved), L1(v/s/i)ROBs, L1(v/s/i)ATs and L1(v/s/i)TLB, in an insulated Golang file called `tile.go`.

Outside the mesh, we remove the L2 Caches and peripheral DRAM memory banks compared to v1. `ToMesh` is a new variable name to replace the `periphPorts`, and we separated these ports in edge of the mesh, i.e., tile[0, 0..width], instead of directly connecting them to mesh in tile[0, 0]. The detailed descriptions of the connections can be found in `gpu.go` comments.

### Build

We update a clean way to use Golang module management in `go.mod` instead of raw codebase as v1 build, thus we could run `go build` to automatically fetch and import the packages for our project.

```bash
git clone git@github.com:sarchlab/akkalat.git
cd akkalat
git checkout cu-level-without-caches
cd samples/fir
go build
./fir -verify -timing
```

## CU-level mesh based on Shader Array (v1)

### Synopsis

In this part, we abstract layers called **mesh** and **tile** between the level GPU and Shader Array.

A **mesh** contains tileWidth \* tileHeight tiles (4\*4 by default, mutable in gpu builder). And we implement each **tile** with a shader array (contains 4 CUs by default, each CU contains 1 L1 Caches and 1 L1 TLB).

Outside the mesh, we lay out the CP, L2 Caches, L2 TLB, etc. `periphPorts` is used to connect ports to mesh in tile[0, 0], and `periphConn` is to directly connect the components which completely insulates from mesh. The detailed descriptions of the connections can be found in gpu.go comments.

Thus, by default the number of main components like shader arrays, CUs, TLBs and Caches is exactly the same with R9 Nano GPU.

### Build

```bash
MGPUSIM_DIR=/path/to/your/mgpusim
# Please be careful about the potential overriding operation,
# although `-b` options would backup the old files in runner
mv -b -v runner/* $MGPUSIM_DIR/samples/runner
cd $MGPUSIM_DIR/fft
go build && ./fft -timing -report-all -parallel -verify
```
