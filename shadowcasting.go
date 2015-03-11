//the source code from http://www.roguebasin.com/index.php?title=Python_shadowcasting_implementation
// translator: Man ZhiYong

package main

import (
	"fmt"
	"github.com/gbin/goncurses"
	"os"
	"os/signal"
)

var FOV_RADIUS int = 10
var m *Map = &Map{}

var mult [][]int = [][]int{
	[]int{1, 0, 0, -1, -1, 0, 0, 1},
	[]int{0, 1, -1, 0, 0, -1, 1, 0},
	[]int{0, 1, 1, 0, 0, -1, -1, 0},
	[]int{1, 0, 0, 1, -1, 0, 0, -1}}

var dungeon []string = []string{
	"###########################################################",
	"#...........#.............................................#",
	"#...........#........#....................................#",
	"#.....................#...................................#",
	"#....####..............#..................................#",
	"#.......#.......................#####################.....#",
	"#.......#...........................................#.....#",
	"#.......#...........##..............................#.....#",
	"#####........#......##..........##################..#.....#",
	"#...#...........................#................#..#.....#",
	"#...#............#..............#................#..#.....#",
	"#...............................#..###############..#.....#",
	"#...............................#...................#.....#",
	"#...............................#...................#.....#",
	"#...............................#####################.....#",
	"#.........................................................#",
	"#.........................................................#",
	"###########################################################"}

type Map struct {
	data   []string
	width  int
	height int
	light  [][]int
	flag   int
}

func (self *Map) init(m []string) {
	self.data = m
	self.width, self.height = len(m[0]), len(m)
	self.light = [][]int{}
	for i := 0; i < self.height; i++ {
		line := []int{}
		for j := 0; j < self.width; j++ {
			line = append(line, 0)
		}
		self.light = append(self.light, line)
	}
	self.flag = 0
}

func (self *Map) square(x, y int) goncurses.Char {
	return goncurses.Char(self.data[y][x])
}

func (self *Map) blocked(x, y int) bool {
	res := false
	if x < 0 || y < 0 || x >= self.width || y >= self.height || self.data[y][x:x+1] == "#" {
		res = true
	}
	return res
}

func (self *Map) lit(x, y int) bool {
	res := false
	if self.light[y][x] == self.flag {
		res = true
	}
	return res
}

func (self *Map) set_lit(x, y int) {
	if 0 <= x && x < self.width && 0 <= y && y < self.height {
		self.light[y][x] = self.flag
	}
}

func (self *Map) _cast_light(cx, cy, row int, start, end float64, radius, xx, xy, yx, yy, id int) {
	//"Recursive lightcasting function"
	var new_start float64
	if start < end {
		return
	}
	radius_squared := radius * radius
	for j := row; j < radius+1; j++ {
		dx, dy := -j-1, -j
		blocked := false
		for {
			dx += 1
			// Translate the dx, dy coordinates into map coordinates:
			X, Y := cx+dx*xx+dy*xy, cy+dx*yx+dy*yy
			//# l_slope and r_slope store the slopes of the left and right
			// extremities of the square we're considering:
			l_slope, r_slope := (float64(dx)-0.5)/(float64(dy)+0.5), (float64(dx)+0.5)/(float64(dy)-0.5)
			if start < r_slope {
				continue
			} else if end > l_slope {
				break
			} else {
				// Our light beam is touching this square; light it:
				if dx*dx+dy*dy < radius_squared {
					self.set_lit(X, Y)
				}
				if blocked {
					// we're scanning a row of blocked squares:
					if self.blocked(X, Y) {
						new_start = r_slope
						continue
					} else {
						blocked = false
						start = new_start
					}
				} else {
					if self.blocked(X, Y) && j < radius {
						// This is a blocking square, start a child scan:
						blocked = true
						self._cast_light(cx, cy, j+1, start, l_slope, radius, xx, xy, yx, yy, id+1)
						new_start = r_slope
					}
				}
			}
			if dx > 0 {
				break
			}
		}
		//# Row is scanned; do next row unless last square was blocked:
		if blocked {
			break
		}
	}
}

func (self *Map) do_fov(x, y, radius int) {
	//"Calculate lit squares from the given location and radius"
	self.flag += 1
	for oct := 0; oct < 8; oct++ {
		self._cast_light(x, y, 1, 1.0, 0.0, radius,
			mult[0][oct], mult[1][oct],
			mult[2][oct], mult[3][oct], 0)
	}
}

func (self *Map) display(s *goncurses.Window, X int, Y int) {
	var ch goncurses.Char
	var attr goncurses.Char
	dark, lit := goncurses.ColorPair(8), goncurses.ColorPair(7)|goncurses.A_BOLD
	for x := 0; x < self.width; x++ {
		for y := 0; y < self.height; y++ {
			if self.lit(x, y) {
				attr = lit
			} else {
				attr = dark
			}
			if x == X && y == Y {
				ch = '@'
				attr = lit
			} else {
				ch = self.square(x, y)
			}
			s.AttrSet(attr)
			s.MoveAddChar(y, x, ch)
		}
	}
	s.Refresh()
}

func color_pairs() []goncurses.Char {
	c := []goncurses.Char{}
	for i := 1; i < 16; i++ {
		j := int16(i)
		//		fmt.Println(j)
		goncurses.InitPair(j, j%8, 0)
		if i < 8 {
			c = append(c, goncurses.ColorPair(j))
		} else {
			c = append(c, goncurses.ColorPair(j)|goncurses.A_BOLD)
		}
	}
	return c
}

func exit(s *goncurses.Window) {
	s.Keypad(false)
	goncurses.Echo(true)
	goncurses.CBreak(false)
	goncurses.End()
	fmt.Println("Normal termination.")
}

func main() {
	s, _ := goncurses.Init()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			exit(s)
			os.Exit(0)
		}
	}()
	goncurses.StartColor()
	goncurses.Echo(false)
	goncurses.CBreak(true)
	color_pairs()
	s.Keypad(true)
	defer func() {
		exit(s)
	}()
	x, y := 36, 13
	m.init(dungeon)
	for {
		m.do_fov(x, y, FOV_RADIUS)
		m.display(s, x, y)
		k := s.GetChar()
		if k == 259 {
			y -= 1
		} else if k == 258 {
			y += 1
		} else if k == 260 {
			x -= 1
		} else if k == 261 {
			x += 1
		} else {
			break
		}
	}
}
