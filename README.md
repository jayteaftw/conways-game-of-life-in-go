# Conways Game of Life in Go: An Analysis of Concurrent Programming

This repository is an examination of how Go's concurrency can be used to increase the perfoamnce of an application. 
To demonstrate this, the program simulates Conway's Game of Life as its workload.

### Background

Conway's Game of Life is 2 dimensional zero-player game which takes place over a predetermined time t abd has cells that have two possible states: dead or alive. The rules of the game can be condesed as such
1. Any live cell with two or three live neighbours survives.
2. Any dead cell with three live neighbours becomes a live cell.
3. All other live cells die in the next generation. Similarly, all other dead cells stay dead.


This is represented as our compute function where grid represents the current state and buffer represents the next state
```
func compute(x int, y int) {
	buffer[x][y] = 0
	n := (grid[x-1][y-1] + grid[x-1][y+0] + grid[x-1][y+1] +
		grid[x+0][y-1] + grid[x+0][y+1] + grid[x+1][y-1] +
		grid[x+1][y+0] + grid[x+1][y+1])
	if grid[x][y] == 0 && n == 3 {
		buffer[x][y] = 1
	} else if n > 3 || n < 2 {
		buffer[x][y] = 0
	} else {
		buffer[x][y] = grid[x][y]
	}
}
```
#### Sequential Update
The sequential update does as it sounds: it iterates over each cell in the grid and updates accordingly. We will treat this as our baseline.
```
func update() error {
	for x := 1; x < WIDTH-1; x++ {
		for y := 1; y < HEIGHT-1; y++ {
			compute(x, y)
		}
	}
	temp := buffer
	buffer = grid
	grid = temp
	return nil
}

```
#### Concurrent Update
The concurrent algorithm is comprised of a simple load balancer model where work is created asynchronous [ func sendFrameOfWork() ] and sent to asynchronous worker routines [func worker(in <-chan int, out chan<- int)], which are initialized when the program first started, that compute said work and update the grid accordingly. 

```
func sendFrameOfWork(in chan<- int) {
	for x := 1; x < WIDTH-1; x++ {
		in <- x
	}
}

func worker(in <-chan int, out chan<- int) {
	for x := range in {
		for y := 1; y < HEIGHT-1; y++ {
			compute(x, y)
		}
		out <- HEIGHT - 2
	}
}

```

Communication is done using 2 buffered channels, in and out: in describes the work that needs to be done and out describes the work that has been done.
```
var in chan int = make(chan int, 2500)
var out chan int = make(chan int, 2500)
```

To signal that computation is done, a syncrounous recieve() function is used to tally all the done calls in the out channel.
```
func recieve(done <-chan int) error {
	count := 0
	total := (HEIGHT - 2) * (WIDTH - 2)
	for {
		if count >= total {
			grid = buffer
			return nil
		}
		count += <-done
	}
}
``

Finally instead of Update() being called, updateConcurrent() is called to begin the process of computing the next time step.
```
func updateConcurrent() error {

	go sendFrameOfWork(in)
	return recieve(out)

}
```


### Experiment

Each simulation iterated 100 times

| Grid (seconds) | Sequential | Concurrent (4 Gophers) | Concurrent (8 Gophers) | Concurrent (16 Gophers) |
| -------------- | ---------- | ---------------------- | ---------------------- | ----------------------- |
| 100x100        | 1.815      | 1.857                  | 1.822                  | 1.806                   |
| 500x500        | 1.817      | 1.836                  | 1.843                  | 1.831                   |
| 1000x1000      | 1.871      | 1.892                  | 1.870                  | 1.811                   |
| 1500x1500      | 3.647      | 1.911                  | 1.859                  | 1.894                   |
| 2000x2000      | 6.277      | 3.742                  | 3.420                  | 3.178                   |
| 2500x2500      | 12.839     | 10.640                 | 10.227                 | 9.954                   |
| 3000x3000      | 18.373     | 15.162                 | 14.549                 | 14.648                  |
| 5000x5000      | 48.841     | 40.166                 | 38.517                 | 38.194                  |



| Grid (performance) | 4 Goroutines | 8 Goroutines | 16 Goroutines |
| ------------------ | ------------ | ------------ | ------------- |
| 100x100            | -0.023       | -0.004       | 0.005         |
| 500x500            | -0.010       | -0.014       | -0.008        |
| 1000x1000          | -0.012       | 0.001        | 0.032         |
| 1500x1500          | 0.476        | 0.490        | 0.481         |
| 2000x2000          | 0.404        | 0.455        | 0.494         |
| 2500x2500          | 0.171        | 0.203        | 0.225         |
| 3000x3000          | 0.175        | 0.208        | 0.203         |
| 5000x5000          | 0.178        | 0.211        | 0.218         |
