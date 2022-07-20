package main

import (
	"errors"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten"
)

const scale = 2
const WIDTH = 1000
const HEIGHT = 1000

var black color.RGBA = color.RGBA{95, 95, 95, 255}    //95,95,95
var white color.RGBA = color.RGBA{233, 233, 233, 255} //233,233,233
var grid [WIDTH][HEIGHT]uint8 = [WIDTH][HEIGHT]uint8{}
var buffer [WIDTH][HEIGHT]uint8 = [WIDTH][HEIGHT]uint8{}
var counter int = 0
var totalCount int = 0
var parallel bool = false

var NumWorkers int = 8
var in chan int = make(chan int, 2500)
var out chan int = make(chan int, 2500)

// Logic

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

type Work struct {
	x int
}

func worker(in <-chan int, out chan<- int) {
	for x := range in {
		//println(w.x, w.y)
		for y := 1; y < HEIGHT-1; y++ {
			compute(x, y)

		}
		out <- HEIGHT - 2
	}

}

func sendFrameOfWork(in chan<- int) {
	for x := 1; x < WIDTH-1; x++ {
		in <- x
	}
}

func recieve(done <-chan int) error {
	count := 0
	//println("here")

	total := (HEIGHT - 2) * (WIDTH - 2)
	for {
		//println("here")
		//println(count, total)

		if count >= total {

			grid = buffer
			return nil
		}

		count += <-done

	}
}

func updateConcurrent() error {

	go sendFrameOfWork(in)
	return recieve(out)

}

func updatePrallel() error {
	var wg = sync.WaitGroup{}
	for x := 1; x < WIDTH-1; x++ {
		for y := 1; y < HEIGHT-1; y++ {
			wg.Add(1)
			go func(x int, y int) {
				compute(x, y)
				wg.Done()
			}(x, y)
		}
	}
	wg.Wait()
	temp := buffer
	buffer = grid
	grid = temp
	return nil
}

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

// Main
func render(screen *ebiten.Image) {
	screen.Fill(white)
	for x := 0; x < WIDTH; x++ {
		for y := 0; y < HEIGHT; y++ {
			if grid[x][y] > 0 {
				for x1 := 0; x1 < scale; x1++ {
					for y1 := 0; y1 < scale; y1++ {
						screen.Set((x*scale)+x1, (y*scale)+y1, black)
					}
				}
			}
		}
	}
}
func frame(screen *ebiten.Image) error {
	totalCount++
	var err error = nil

	if parallel {
		err = updateConcurrent()
	} else {
		err = update()
	}

	if !ebiten.IsDrawingSkipped() {
		render(screen)
	}

	if totalCount == 100 {
		return errors.New("stop")
	}

	return err
}

func start() {
	if parallel {
		numCPUs := runtime.NumCPU()
		runtime.GOMAXPROCS(numCPUs)
		for i := 0; i < NumWorkers; i++ {
			go worker(in, out)
		}
	}

	for x := 1; x < WIDTH-1; x++ {
		for y := 1; y < HEIGHT-1; y++ {
			if rand.Float32() < 0.5 {
				grid[x][y] = 1
			}
		}
	}
	start := time.Now()
	if err := ebiten.Run(frame, WIDTH*scale, HEIGHT*scale, 2, "Conway's Game of Go"); err != nil {
		t := time.Now()
		finalTime := t.Sub(start)
		fmt.Printf("Finished: %v \n", finalTime)
		log.Fatal(err)
	}

}

func main() {
	parallel = true
	start()
}
