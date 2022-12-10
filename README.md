# SQLite: CGO vs no CGO

This repo benchmarks mattn/go-sqlite3 against modernc.org/sqlite which
is a translation of SQLite3 from C to Go. This translation allows the
latter package to avoid CGO since there is no C.

My initial observations showed it being twice as slow as
mattn/go-sqlite3 and this repo is to test that observation.

See the [blog post for details](https://datastation.multiprocess.io/blog/2022-05-12-sqlite-in-go-with-and-without-cgo.html).


## Fork

This fork differs from the original in the following aspects:

- After querying database rows, iterate over it (scan them). Why? Because without iterating,
  values might not be fetched from the database.

- Add benchmark for https://github.com/cvilsmeier/sqinn-go


## Machine Specs

- OS: x86_64 GNU/Linux (Debian 11.5)
- RAM: 16 GB
- Disk: 1TB NVME SSD
- Processor: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz


## Results

10000 rows:

	                  cgo             nocgo           sqinn-go
	INSERT            0.033           0.060           0.019
	GROUP_BY          0.002           0.003           0.006


479827 rows:

	                  cgo             nocgo           sqinn-go
	INSERT            1.669           2.851           0.840
	GROUP_BY          0.187           0.202           0.230

4798270 rows:

	                  cgo             nocgo           sqinn-go
	INSERT            16.938          28.609           8.504
	GROUP_BY           3.620           3.630           2.153


## Summary

For small database workloads, the library you're using does not
matter so much. For medium to large database workloads however,
you'll be better off with cvilsmeier/sqinn-go.
