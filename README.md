# akkalat

Akkalat is the infrastructure to simulate wafer-scale GPUs, based on the architecture-agnostic simulator framework [Akita](https://gitlab.com/akita/akita).

The simulator targeted on [Cerebras Wafer-Scale Engines](https://cerebras.net/blog/cerebras-wafer-scale-engine-why-we-need-big-chips-for-deep-learning/) and enable OpenCL workloads running as benchmarks.

In the latest stable Akkalat v3, we model a tile-based system at *CU(Compute Unit)-level*. The tile consists of optimized L1 caches and TLBs, distributed memory and other basic components like reorder buffers and address translators.

The further performance optimization and hardware designs could be assisted with [Akita Vis4Mesh](https://github.com/ueqri/vis4mesh), a visualization tool developed later for mesh Network-on-Chip research.

## Build

```bash
git clone git@github.com:ueqri/akkalat.git
cd akkalat/samples/fir
go build

# Case 1: run FIR in 64x64 mesh without NoC tracing
./fir -width=64 -height=64 -length=100000 -verify -timing

# Case 2: run FIR with tracing and Vis4Mesh,
# and customize the env to meet your demands
export AKITA_TRACE_PASSWORD=[RedisPassword]
export AKITA_TRACE_IP=127.0.0.1
export AKITA_TRACE_PORT=6379
export AKITA_TRACE_REDIS_DB=0
./fir -width=64 -height=64 -length=100000 -verify -timing -trace-vis
# Then follow Vis4Mesh tutorial to run visualization
```

Notice: many akita libraries used here are also under development, so make sure the go packages `util`, `noc` and `mgpusim` (as we see in `go.mod`) are all cloned and on the latest commit of correct branches.

```bash
# Assume the work directory is inside akkalat/
cd ..
git clone --single-branch -b v2 git@gitlab.com/akita/util
git clone --single-branch -b 9-task-tracing-for-networks git@gitlab.com/akita/noc
git clone --single-branch -b 90-command-processor-reuses-todma-sender-port git@gitlab.com/akita/mgpusim
```

## Benchmark

We support the same benchmark set as MGPUSim.

| AMD APP SDK           | DNN Mark   | HeteroMark | Polybench | Rodinia          | SHOC      |
| --------------------- | ---------- | ---------- | --------- | ---------------- | --------- |
| Bitonic Sort          | MaxPooling | AES        | ATAX      | Needleman-Wunsch | BFS       |
| Fast Walsh Transform  | ReLU       | FIR        | BICG      |                  | FFT       |
| Floyd-Warshall        |            | KMeans     |           |                  | SPMV      |
| Matrix Multiplication |            | PageRank   |           |                  | Stencil2D |
| Matrix Transpose      |            |            |           |                  |           |
| NBody                 |            |            |           |                  |           |
| Simple Covolution     |            |            |           |                  |           |

### Performance

We demonstrate a 32x32 wafer-scale GPU using akkalat. Compared to 16 unified AMD R9Nano GPUs(with the same numbers of compute units) modeled by [MGPUSim](https://gitlab.com/akita/mgpusim), the wafer-scale outperforms in many workloads and achieves up-to 4 speedup in Polybench.

This figure is generated full-automatically by another self-developed tool [akitaplot](https://github.com/ueqri/akitaplot), feel free to try it :-). And the detailed metrics of this test could also download in that repo, [link](https://github.com/ueqri/akitaplot/blob/main/example_data.tar.gz).

![speedup](https://raw.githubusercontent.com/ueqri/akitaplot/main/example_kernelTime.png)

## Changelog

### Nov 8, 2021

After profiling the v2 model in [Daisen](https://osf.io/73ry8/), we found the **Instruction Caches** and **Scalar Caches** are definitely needed to avoid massive global memory requests. So we attached these specific caches inside each tile.

This is the **stable** version recently, we would do further research based on the visualization results of [Vis4Mesh](https://github.com/ueqri/vis4mesh). Further steps includes applying large-bandwidth DMA engine along the mesh edge, optimizing data placement for better locality and so on.

### Full history

To see the development progress, please switch to `CHANGE.log`, click [here](https://github.com/ueqri/akkalat/blob/main/CHANGELOG.md).
