# monday

## Banchmark commands

```
go test -run=xxx -bench=^BenchmarkActorGenerate$ -benchmem -cpuprofile=cpu.out -memprofile=mem.out
go tool pprof -pdf keygen.test.exe cpu.out > cpu0.pdf
go tool pprof -pdf keygen.test.exe mem.out > mem0.pdf
go tool pprof keygen.test.exe cpu.out
```


## Build

bash:
```
GOARCH=amd64 go install .
```