package game

import (
	"fmt"
	"math/rand"
)

const (
	mapWidth  = 30
	mapHeight = 15
)

type point struct {
	x int
	y int
}

type room struct {
	x int
	y int
	w int
	h int
}

func (r room) center() point {
	return point{x: r.x + r.w/2, y: r.y + r.h/2}
}

// dungeonMap is a generated room-and-corridor map with a route from start to boss.
type dungeonMap struct {
	grid  [][]byte
	path  []point
	start point
	boss  point
}

func generateDungeon(rng *rand.Rand) (dungeonMap, error) {
	for attempt := 0; attempt < 20; attempt++ {
		generated, err := generateDungeonOnce(rng)
		if err == nil {
			return generated, nil
		}
	}
	return dungeonMap{}, fmt.Errorf("could not generate a connected dungeon")
}

func generateDungeonOnce(rng *rand.Rand) (dungeonMap, error) {
	grid := make([][]byte, mapHeight)
	for y := range grid {
		grid[y] = make([]byte, mapWidth)
		for x := range grid[y] {
			grid[y][x] = '#'
		}
	}

	var rooms []room
	for attempts := 0; attempts < 120 && len(rooms) < 7; attempts++ {
		w := 3 + rng.Intn(5)
		h := 2 + rng.Intn(4)
		x := 1 + rng.Intn(mapWidth-w-2)
		y := 1 + rng.Intn(mapHeight-h-2)
		candidate := room{x: x, y: y, w: w, h: h}
		overlaps := false
		for _, existing := range rooms {
			if roomsOverlap(candidate, existing) {
				overlaps = true
				break
			}
		}
		if !overlaps {
			rooms = append(rooms, candidate)
		}
	}
	if len(rooms) < 4 {
		return dungeonMap{}, fmt.Errorf("not enough rooms")
	}

	for _, r := range rooms {
		for y := r.y; y < r.y+r.h; y++ {
			for x := r.x; x < r.x+r.w; x++ {
				grid[y][x] = '.'
			}
		}
	}

	for i := 1; i < len(rooms); i++ {
		carveCorridor(grid, rooms[i-1].center(), rooms[i].center(), rng)
	}

	start := rooms[0].center()
	boss := rooms[len(rooms)-1].center()
	path := findPath(grid, start, boss)
	if len(path) < 8 {
		return dungeonMap{}, fmt.Errorf("route too short")
	}

	return dungeonMap{grid: grid, path: path, start: start, boss: boss}, nil
}

func roomsOverlap(a, b room) bool {
	return a.x <= b.x+b.w+1 &&
		a.x+a.w+1 >= b.x &&
		a.y <= b.y+b.h+1 &&
		a.y+a.h+1 >= b.y
}

func carveCorridor(grid [][]byte, from, to point, rng *rand.Rand) {
	if rng.Intn(2) == 0 {
		carveHorizontal(grid, from.x, to.x, from.y)
		carveVertical(grid, from.y, to.y, to.x)
		return
	}
	carveVertical(grid, from.y, to.y, from.x)
	carveHorizontal(grid, from.x, to.x, to.y)
}

func carveHorizontal(grid [][]byte, x1, x2, y int) {
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	for x := x1; x <= x2; x++ {
		grid[y][x] = '.'
	}
}

func carveVertical(grid [][]byte, y1, y2, x int) {
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	for y := y1; y <= y2; y++ {
		grid[y][x] = '.'
	}
}

func findPath(grid [][]byte, start, target point) []point {
	visited := make([][]bool, len(grid))
	for y := range visited {
		visited[y] = make([]bool, len(grid[y]))
	}
	previous := make(map[point]point)
	queue := []point{start}
	visited[start.y][start.x] = true
	directions := []point{{x: 0, y: -1}, {x: 1, y: 0}, {x: 0, y: 1}, {x: -1, y: 0}}

	found := false
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current == target {
			found = true
			break
		}
		for _, direction := range directions {
			next := point{x: current.x + direction.x, y: current.y + direction.y}
			if next.y < 0 || next.y >= len(grid) || next.x < 0 || next.x >= len(grid[next.y]) {
				continue
			}
			if grid[next.y][next.x] == '#' || visited[next.y][next.x] {
				continue
			}
			visited[next.y][next.x] = true
			previous[next] = current
			queue = append(queue, next)
		}
	}
	if !found {
		return nil
	}

	path := []point{target}
	for path[len(path)-1] != start {
		path = append(path, previous[path[len(path)-1]])
	}
	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
		path[left], path[right] = path[right], path[left]
	}
	return path
}

func (m dungeonMap) pathSet() map[point]bool {
	result := make(map[point]bool, len(m.path))
	for _, p := range m.path {
		result[p] = true
	}
	return result
}

func (m dungeonMap) walkable(p point) bool {
	return p.y >= 0 && p.y < len(m.grid) && p.x >= 0 && p.x < len(m.grid[p.y]) && m.grid[p.y][p.x] != '#'
}

func (m dungeonMap) floorCells() []point {
	points := make([]point, 0, mapWidth*mapHeight/2)
	for y, row := range m.grid {
		for x, cell := range row {
			if cell != '#' {
				points = append(points, point{x: x, y: y})
			}
		}
	}
	return points
}
