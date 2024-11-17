package main

import "math"

const (
	MAZESIZE = 32
	TILESIZE = 300
	MAXDIST  = 300
)

var (
	maze       [32][32]bool
	px, py     int
	mapx, mapy int

	dx, dy float64
)

func init() {

	maze = [32][32]bool{
		[32]bool{true, true, true, true, true, true, true, true, true, true, false, false, false, true, false, true, false, false, false, true, true, true, true, true, true, true, true, true, true, true, true, true},
		[32]bool{true, false, false, false, false, false, false, false, false, true, false, false, false, true, false, true, false, false, false, true, false, false, false, false, false, true, true, false, false, false, false, true},
		[32]bool{true, false, true, true, true, false, true, true, false, true, false, false, false, true, false, true, false, false, false, true, false, true, true, true, false, true, true, false, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, false, true, false, false, false, true, false, true, false, false, false, true, false, true, true, true, false, false, false, false, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, false, true, false, false, false, true, false, true, false, false, false, true, false, true, true, true, true, true, true, false, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, false, true, true, true, true, true, false, true, true, true, true, true, false, true, true, true, true, true, true, false, true, true, false, true},
		[32]bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, false, true, true, false, true, true, true, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, false, true, true, false, true, true, true, true, true, false, true},
		[32]bool{true, false, true, true, false, false, false, false, false, true, true, false, false, false, false, false, false, false, false, false, false, false, true, true, false, false, false, false, true, true, false, true},
		[32]bool{true, false, true, true, false, true, true, true, false, true, true, false, true, true, true, true, true, true, false, true, true, false, true, true, false, true, true, false, true, true, false, true},
		[32]bool{true, false, true, true, false, true, true, true, false, true, true, false, true, false, false, false, false, true, false, true, true, false, true, true, false, true, true, false, true, true, false, true},
		[32]bool{true, false, false, false, false, false, true, true, false, true, true, false, true, false, false, false, false, true, false, true, true, false, true, true, false, true, true, false, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, false, false, false, false, true, false, false, false, false, true, false, true, true, false, true, true, false, true, true, false, true, true, false, true},
		[32]bool{false, false, true, true, true, false, true, true, true, true, true, false, true, false, false, false, false, true, false, true, true, false, false, false, false, true, true, false, false, false, false, false},
		[32]bool{true, false, true, true, true, false, true, true, true, true, true, false, true, false, false, false, false, true, false, true, true, true, true, true, false, true, true, true, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, true, true, true, false, true, false, false, false, false, true, false, true, true, true, true, true, false, true, true, true, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, false, false, false, false, true, false, false, false, false, true, false, true, true, false, false, false, false, true, true, false, false, false, false, true},
		[32]bool{true, false, true, true, true, false, true, true, false, true, true, false, true, false, false, false, false, true, false, true, true, false, true, true, false, true, true, false, true, true, false, true},
		[32]bool{true, false, false, false, false, false, true, true, false, true, true, false, true, false, false, false, false, true, false, true, true, false, true, true, false, true, true, false, true, true, false, true},
		[32]bool{true, false, true, true, false, true, true, true, false, true, true, false, true, false, false, false, false, true, false, true, true, false, true, true, false, true, true, false, true, true, false, true},
		[32]bool{true, false, true, true, false, true, true, true, false, true, true, false, true, true, true, true, true, true, false, true, true, false, true, true, false, true, true, false, true, true, false, true},
		[32]bool{true, false, true, true, false, false, false, false, false, true, true, false, false, false, false, false, false, false, false, false, false, false, true, true, false, false, false, false, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, false, true, true, false, true, true, true, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, true, true, true, true, true, true, false, true, true, true, true, true, true, false, true, true, false, true, true, true, true, true, false, true},
		[32]bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, false, true, true, true, true, true, false, true, true, true, true, true, false, true, true, true, true, true, true, false, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, false, true, false, false, false, true, false, true, false, false, false, true, false, true, true, true, true, true, true, false, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, false, true, false, false, false, true, false, true, false, false, false, true, false, true, true, true, false, false, false, false, true, true, false, true},
		[32]bool{true, false, true, true, true, false, true, true, false, true, false, false, false, true, false, true, false, false, false, true, false, true, true, true, false, true, true, false, true, true, false, true},
		[32]bool{true, false, false, false, false, false, false, false, false, true, false, false, false, true, false, true, false, false, false, true, false, false, false, false, false, true, true, false, false, false, false, true},
		[32]bool{true, true, true, true, true, true, true, true, true, true, false, false, false, true, false, true, false, false, false, true, true, true, true, true, true, true, true, true, true, true, true, true},
	}

}

func castRay(rayAngle float64) int {
	dx = math.Cos(rayAngle)
	dy = math.Sin(rayAngle)

	for i := 1; i < MAXDIST; i += 2 {
		mapx = (px + int(dx*float64(i))) / TILESIZE
		mapy = (py + int(dy*float64(i))) / TILESIZE

		if mapx < 0 || mapy < 0 || mapx >= MAZESIZE || mapy >= MAZESIZE {
			return i
		}

		if maze[mapy][mapx] {
			return i
		}
	}
	return MAXDIST
}

func printTile(x, y int) {

	tx := x / 300
	ty := y / 300
	println(x, y, tx, ty)

	x = (x % 300) / 30
	y = (y % 300) / 30

	println(x, y, tx, ty)
	for j := -1; j <= 1; j++ {
		for i := -1; i <= 1; i++ {
			if ty+j < 0 || ty+j > 31 ||
				tx+i < 0 || tx+i > 31 {
				print(" ")
			} else if maze[ty+j][tx+i] {
				print("#")
			} else {
				print(" ")
			}
		}
		println("")
	}

	println("+----------+")
	for j := 0; j < 10; j++ {
		print("|")
		for i := 0; i < 10; i++ {
			if x == i && y == j {
				print("X")
			} else {
				print(" ")
			}
		}
		println("|")
	}
	println("+----------+")

}
