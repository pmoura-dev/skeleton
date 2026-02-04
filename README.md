# Skeleton

**Skeleton** is a small Go library for building production-ready services with a structured lifecycle and built-in operation features.

It provides a simple framework to:
* run multiple long-lived tasks (goroutines)
* handle OS signals and fatal task errors
* perform coordinated graceful shutdown
* expost a clear service lifecycle
* support health, readiness, and observability out of the box

It acts as a foundation for Go applications.


## Basic Usage

```go
func main() {
  app := skeleton.New(skeleton.Config{})

  app.AddTask(httpServerTask)
  app.AddTask(workerTask)

  if err := app.Run(); err != nil {
    os.Exit(1)
  }
}
```

Skeleton will:
1. Start all registered tasks
2. Wait for either:
   - an OS shutdown signal, or
   - a fatal task error
3. Initiate graceful shutdown
4. Wait for all tasks to stop (or timeout)
5. Exit with an appropriate status

