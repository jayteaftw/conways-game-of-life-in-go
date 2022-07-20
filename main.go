package main

import (
	"errors"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten"
)

const scale = 2
const WIDTH = 500
const HEIGHT = 500

var black color.RGBA = color.RGBA{95, 95, 95, 255}    //95,95,95
var white color.RGBA = color.RGBA{233, 233, 233, 255} //233,233,233
var grid [WIDTH][HEIGHT]uint8 = [WIDTH][HEIGHT]uint8{}
var buffer [WIDTH][HEIGHT]uint8 = [WIDTH][HEIGHT]uint8{}
var counter int = 0
var totalCount int = 0
var parallel bool = false

// Logic

func updatePrallel() error {
	var wg = sync.WaitGroup{}
	for x := 1; x < WIDTH-1; x++ {
		for y := 1; y < HEIGHT-1; y++ {
			wg.Add(1)
			go func(x int, y int) {
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
	counter++
	totalCount++
	var err error = nil
	if counter == 1 {
		if parallel {
			err = updatePrallel()
		} else {
			err = update()
		}
		counter = 0
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
		finalTime := time.Duration(t.Sub(start).Microseconds())
		fmt.Printf("Finished: %v \n", finalTime)
		log.Fatal(err)
	}

}

func main() {
	parallel = true
	start()
}
