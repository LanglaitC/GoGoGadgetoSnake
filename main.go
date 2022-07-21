package main

import (
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
	"math/rand"
	"strconv"
	"time"
)

type BlockType int
type Move int
type GameState int

const (
	Up Move = iota
	Right
	Down
	Left
	Escape
	NoDirection
)

const BOARD_WIDTH = 40
const BOARD_HEIGHT = 20
const GAME_SPEED = 120

const (
	Wall BlockType = iota
	Apple
	Snake
	Nothing
)

const (
	Pending GameState = iota
	BorderFinish
	SnakeFinish
	QuitByUser
)

type Board struct {
	width  int
	height int
	middle Point
	yStart int
	xStart int
	yEnd   int
	xEnd   int
}

type Game struct {
	Snake  Block
	Board  Board
	Apple  Point
	isOver bool
	state  GameState
	score  int
}

type Block struct {
	position Point
	next     *Block
}

type Point struct {
	x int
	y int
}

func handleKeyboard(c1 chan Move) {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowLeft:
				c1 <- Left
			case termbox.KeyArrowDown:
				c1 <- Down
			case termbox.KeyArrowRight:
				c1 <- Right
			case termbox.KeyArrowUp:
				c1 <- Up
			case termbox.KeyEsc:
				c1 <- Escape
			default:
				c1 <- NoDirection
			}
		case termbox.EventError:
			panic("Key capture went wrong")
		}
	}
}

func initGame() Game {
	width, height := termbox.Size()
	middleX := width / 2
	middleY := height / 2
	xStart := middleX - BOARD_WIDTH/2
	xEnd := middleX + BOARD_WIDTH/2
	yStart := middleY - BOARD_HEIGHT/2
	yEnd := middleY + BOARD_HEIGHT/2
	game := Game{
		Snake: Block{
			position: Point{
				x: middleX,
				y: middleY,
			},
			next: &Block{
				position: Point{
					x: middleX + 1,
					y: middleY,
				},
				next: &Block{
					position: Point{
						x: middleX + 2,
						y: middleY,
					},
				},
			},
		},
		Board: Board{
			width:  width,
			height: height,
			middle: Point{
				x: middleX,
				y: middleY,
			},
			xStart: xStart,
			yStart: yStart,
			yEnd:   yEnd,
			xEnd:   xEnd,
		},
		Apple: Point{
			x: -1,
			y: -1,
		},
		score: 0,
	}
	return game
}

func (game *Game) renderArena() {
	for i := game.Board.yStart; i < game.Board.yEnd; i++ {
		for h := game.Board.xStart; h < game.Board.xEnd; h++ {
			if game.isWall(h, i) {
				termbox.SetCell(h, i, ' ', termbox.ColorRed, 0xFF5500)
			} else {
				termbox.SetCell(h, i, ' ', termbox.ColorCyan, termbox.ColorCyan)
			}
		}
	}
}

func (game *Game) renderSnake() {
	tmp := game.Snake
	for tmp.next != nil {
		termbox.SetCell(tmp.position.x, tmp.position.y, ' ', termbox.ColorBlack, termbox.ColorBlack)
		tmp = *tmp.next
	}
	termbox.SetCell(tmp.position.x, tmp.position.y, ' ', termbox.ColorBlack, termbox.ColorBlack)
}

func (game *Game) renderApple() {
	termbox.SetCell(game.Apple.x, game.Apple.y, ' ', termbox.ColorDarkGray, termbox.ColorDarkGray)
}

func (game *Game) renderScore() {
	msg := "Score: " + strconv.Itoa(game.score)
	x := game.Board.xStart
	y := game.Board.yEnd + 1
	for _, c := range msg {
		termbox.SetCell(x, y, c, termbox.ColorBlack, termbox.ColorLightGreen)
		x += runewidth.RuneWidth(c)
	}
}

func (game *Game) render() {
	//termbox.SetOutputMode(termbox.OutputRGB)
	termbox.Clear(termbox.ColorLightGreen, termbox.ColorLightGreen)
	game.renderArena()
	game.renderSnake()
	game.renderApple()
	game.renderScore()
	termbox.Flush()
}

func (snake *Block) addOneBlock() {
	tmp := snake
	for tmp.next != nil {
		tmp = tmp.next
	}
	tmp.next = &Block{
		position: Point{
			x: -1,
			y: -1,
		},
		next: nil,
	}
}

func (game *Game) makeSnakeMove(currentDirection Move) {
	previousPosition := game.Snake.position
	if currentDirection == Left {
		game.Snake.position.x -= 1
	} else if currentDirection == Up {
		game.Snake.position.y -= 1
	} else if currentDirection == Right {
		game.Snake.position.x += 1
	} else if currentDirection == Down {
		game.Snake.position.y += 1
	}
	tmp := &game.Snake
	for tmp.next != nil {
		tmp = tmp.next
		if tmp.position == game.Snake.position {
			game.state = SnakeFinish
			game.isOver = true
		}
		tmpPos := tmp.position
		tmp.position = previousPosition
		previousPosition = tmpPos

	}
	if game.Snake.position == game.Apple {
		game.Snake.addOneBlock()
		game.score += 1
		game.createApple()
	}
	if game.isWall(game.Snake.position.x, game.Snake.position.y) {
		game.state = BorderFinish
		game.isOver = true
	}
}

func (game *Game) isWall(x int, y int) bool {
	return y == game.Board.yStart || x == game.Board.xStart || y == game.Board.yEnd-1 || x == game.Board.xEnd-1
}

func (game *Game) getSnakePositions() map[Point]bool {
	snakePositions := map[Point]bool{}
	tmp := &game.Snake
	for tmp.next != nil {
		snakePositions[tmp.position] = true
		tmp = tmp.next
	}
	return snakePositions
}

func (game *Game) createApple() {
	snakePositions := game.getSnakePositions()
	availablePositions := make([]Point, 0)
	for i := game.Board.yStart; i < game.Board.yEnd; i++ {
		for h := game.Board.xStart; h < game.Board.xEnd; h++ {
			if !game.isWall(h, i) && !snakePositions[Point{x: h, y: i}] {
				availablePositions = append(availablePositions, Point{x: h, y: i})
			}
		}
	}
	game.Apple = availablePositions[rand.Intn(len(availablePositions))]
}

func (game *Game) commandIsValid(move Move) bool {
	if move == Left && game.Snake.next.position.x == game.Snake.position.x-1 {
		return false
	} else if move == Right && game.Snake.next.position.x == game.Snake.position.x+1 {
		return false
	} else if move == Down && game.Snake.next.position.y == game.Snake.position.y+1 {
		return false
	} else if move == Up && game.Snake.next.position.y == game.Snake.position.y-1 {
		return false
	}
	return true
}

func main() {
	termbox.Init()
	game := initGame()
	currentDirection := Left
	c1 := make(chan Move)
	go handleKeyboard(c1)
	game.createApple()
	for {
		select {
		case command := <-c1:
			{
				if command == Escape {
					game.isOver = true
					game.state = QuitByUser
				} else if command != NoDirection && game.commandIsValid(command) {
					currentDirection = command
				}
			}
		default:
		}
		game.makeSnakeMove(currentDirection)
		game.render()
		time.Sleep(GAME_SPEED * time.Millisecond)
		if game.isOver == true {
			break
		}
	}
	termbox.Close()
}
