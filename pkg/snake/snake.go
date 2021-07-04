package snake

import (
	"github.com/BattlesnakeOfficial/rules"
	"strconv"
	"log"
	"os"
)

type Coord struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

type Snake struct {
	Id      string  `json:"id"`
	Name    string  `json:"name"`
	Health  int32   `json:"health"`
	Body    []Coord `json:"body"`
	Latency string  `json:"latency"`
	Head    Coord   `json:"head"`
	Length  int32   `json:"length"`
	Shout   string  `json:"shout"`
	Squad   string  `json:"squad"`
}

type Board struct {
	Height  int32   `json:"height"`
	Width   int32   `json:"width"`
	Food    []Coord `json:"food"`
	Hazards []Coord `json:"hazards"`
	Snakes  []Snake `json:"snakes"`
}

type Ruleset struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Game struct {
	Id      string  `json:"id"`
	Timeout int32   `json:"timeout"`
	Ruleset Ruleset `json:"ruleset"`
}

type Payload struct {
	Game  Game  `json:"game"`
	Turn  int32 `json:"turn"`
	Board Board `json:"board"`
	You   Snake `json:"you"`
}

type Response struct {
	Move  string `json:"move"`
	Shout string `json:"shout"`
}

type Node struct {
	Move     rules.SnakeMove
	Value    int
	Children []Node
}

func Move(p Payload) string {
	var ruleset rules.Ruleset
	standard := rules.StandardRuleset{
		FoodSpawnChance: 15,
		MinimumFood:     1,
	}
	ruleset = &standard
	boardState := getBoardStateFromBoard(p.Board)
	boardMap := buildBoardMap(p)
	possibleMoves := [4]Node{
		Node{
			Move:     rules.SnakeMove{ID: p.You.Id, Move: "up"},
			Value:    -1000000,
			Children: make([]Node, 0),
		},
		Node{
			Move:     rules.SnakeMove{ID: p.You.Id, Move: "down"},
			Value:    -1000000,
			Children: make([]Node, 0),
		},
		Node{
			Move:     rules.SnakeMove{ID: p.You.Id, Move: "left"},
			Value:    -1000000,
			Children: make([]Node, 0),
		},
		Node{
			Move:     rules.SnakeMove{ID: p.You.Id, Move: "right"},
			Value:    -1000000,
			Children: make([]Node, 0),
		},
	}
	move := "left"
	value := -2000000
	for _, n := range possibleMoves {
		val := alphaBeta(n, 15, -1000000, 1000000, true, p.You.Id, p.You.Id, ruleset, boardState, make([]rules.SnakeMove, 0))
		if val > value {
			move = n.Move.Move
			value = val
		}
	}
	if value <= -1000000 {
		move = findBestAdjacent(p.You.Head, boardMap)
	}
	return move
}

func alphaBeta(node Node, depth int, alpha int, beta int, maximizingPlayer bool, youID string, currentID string, r rules.Ruleset, b *rules.BoardState,
	thisTurnMoves []rules.SnakeMove) int {
	if b == nil || b.Snakes == nil {
		if isDebug() {
			log.Printf("-> %d %s %s %d (%d/%d) ^^^\n", depth, node.Move.Move, currentID, -1000002, alpha, beta)
		}
		return -1000002
	}
	thisValue := scoreMoveOnBoardState(youID, node.Move, r, b)
	gameIsOver, _ := r.IsGameOver(b)
	if depth == 0 || gameIsOver || thisValue <= -1000000 {
		node.Value = thisValue
		if isDebug() {
			log.Printf("-> %d %s %s %d (%d/%d) !!!\n", depth, node.Move.Move, currentID, thisValue, alpha, beta)
		}
		return thisValue
	}
	possibleMoves := []rules.SnakeMove{
		rules.SnakeMove{ID: currentID, Move: "up"},
		rules.SnakeMove{ID: currentID, Move: "down"},
		rules.SnakeMove{ID: currentID, Move: "left"},
		rules.SnakeMove{ID: currentID, Move: "right"},
	}
	movesToDelete := make([]int, 0)
	for i, m := range possibleMoves {
		value := scoreMoveOnBoardState(youID, m, r, b)
		if value <= -1000000 {
			movesToDelete = append(movesToDelete, i)
		}
	}
	for _, i := range movesToDelete {
		possibleMoves[i].Move = "delete"
	}
	nextID := youID
	lastWasCurrent := false
	isLastThisRound := false
	copyTurnMoves := make([]rules.SnakeMove, len(thisTurnMoves))
	for i, v := range thisTurnMoves {
		copyTurnMoves[i] = v
	}
	for i, s := range b.Snakes {
		if i == (len(b.Snakes) - 1) {
			isLastThisRound = true
		}
		if lastWasCurrent {
			nextID = s.ID
			break
		} else if currentID == s.ID {
			lastWasCurrent = true
		}
	}
	if maximizingPlayer {
		value := -1000000
		for _, m := range possibleMoves {
			if m.Move == "delete" {
				continue
			}
			node.Children = append(node.Children, Node{
				Move:     m,
				Children: make([]Node, 0),
			})
		}
		for _, n := range node.Children {
			copyTurnMoves = append(copyTurnMoves, n.Move)
			value = max(value, alphaBeta(n, depth-1, alpha, beta,
				false, youID, nextID, r, b,
				copyTurnMoves))
			alpha = max(alpha, value)
			if value >= beta {
				break
			}
		}
		if isDebug() {
			log.Printf("-> %d %s %s %d (%d/%d)\n", depth, node.Move.Move, currentID, value, alpha, beta)
		}
		return value
	} else {
		value := 1000000
		for _, m := range possibleMoves {
			if m.Move == "delete" {
				continue
			}
			node.Children = append(node.Children, Node{
				Move:     m,
				Children: make([]Node, 0),
			})
		}
		for _, n := range node.Children {
			copyTurnMoves = append(copyTurnMoves, n.Move)
			boardState := b
			if isLastThisRound {
				boardState, _ = r.CreateNextBoardState(b,
					copyTurnMoves)
			}
			value = min(value, alphaBeta(n, depth-1, alpha, beta,
				isLastThisRound, youID, nextID, r,
				boardState, copyTurnMoves))
			beta = min(beta, value)
			if value <= alpha {
				break
			}
		}
		if isDebug() {
			log.Printf("-> %d %s %s %d (%d/%d)\n", depth, node.Move.Move, currentID, value, alpha, beta)
		}
		return value
	}
}

func scoreMoveOnBoardState(youID string, m rules.SnakeMove, r rules.Ruleset, b *rules.BoardState) int {
	if b == nil || b.Snakes == nil {
		return -1000003
	}
	moves := make([]rules.SnakeMove, 0)
	moves = append(moves, m)
	safeMoveMap := make(map[string]map[string]bool, 0)
	for _, s := range b.Snakes {
		safeMoves := map[string]bool{
			"up":    true,
			"down":  true,
			"left":  true,
			"right": true,
		}
		sHead := s.Body[0]
		// Don't Run Into Bodies
		for _, os := range b.Snakes {
			for _, sb := range os.Body {
				if (sHead.Y+1) == sb.Y && sHead.X == sb.X {
					if _, ok := safeMoves["up"]; ok {
						delete(safeMoves, "up")
					}
				}
				if (sHead.Y-1) == sb.Y && sHead.X == sb.X {
					if _, ok := safeMoves["down"]; ok {
						delete(safeMoves, "down")
					}
				}
				if (sHead.X-1) == sb.X && sHead.Y == sb.Y {
					if _, ok := safeMoves["left"]; ok {
						delete(safeMoves, "left")
					}
				}
				if (sHead.X+1) == sb.X && sHead.Y == sb.Y {
					if _, ok := safeMoves["right"]; ok {
						delete(safeMoves, "right")
					}
				}
			}
		}
		// Don't Hit Walls
		if sHead.Y >= (b.Height - 1) {
			if _, ok := safeMoves["up"]; ok {
				delete(safeMoves, "up")
			}
		}
		if sHead.Y <= 0 {
			if _, ok := safeMoves["down"]; ok {
				delete(safeMoves, "down")
			}
		}
		if sHead.X <= 0 {
			if _, ok := safeMoves["left"]; ok {
				delete(safeMoves, "left")
			}
		}
		if sHead.X >= (b.Width - 1) {
			if _, ok := safeMoves["right"]; ok {
				delete(safeMoves, "right")
			}
		}
		// Pick from What's Left
		move := "left"
		for k, _ := range safeMoves {
			move = k
			break
		}
		safeMoveMap[s.ID] = safeMoves
		if m.ID == s.ID {
			continue
		}
		moves = append(moves, rules.SnakeMove{ID: s.ID, Move: move})
	}
	nextBoard, _ := r.CreateNextBoardState(b, moves)
	score := 0
	// Find Yourself
	var you rules.Snake
	for _, s := range nextBoard.Snakes {
		if s.ID == youID {
			you = s
			yHead := you.Body[0]
			if len(s.EliminatedCause) > 0 {
				return -1000001
			}
			if yHead.X < 0 || yHead.X >= nextBoard.Width || yHead.Y < 0 || yHead.Y >= nextBoard.Height {
				return -1000001
			}
			if len(safeMoveMap[s.ID]) <= 0 {
				score = score - 1000
			}
			if s.Health < 20 {
				score = score - 100
			}
			for _, f := range b.Food {
				dist := abs(yHead.X-f.X) + abs(yHead.Y-f.Y)
				if dist == 0 {
					dist = 1
				}
				foodScore := int((1.0 / float32(dist)) * 5000)
				score = score + foodScore
			}
			for _, os := range b.Snakes {
				for i, osb := range os.Body {
					if i != 0 && yHead.X == osb.X && yHead.Y == osb.Y {
						return -1000001
					}
				}
				osTail := os.Body[len(os.Body)-1]
				osTail2 := os.Body[len(os.Body)-2]
				osHas2Tail := osTail.X == osTail2.X && osTail.Y == osTail2.Y
				dist := abs(yHead.X-osTail.X) + abs(yHead.Y-osTail.Y)
				if !osHas2Tail {
					tailScore := int((1.0 / float32(dist)) * 1000)
					score = score + tailScore
				}
			}
			break
		}
	}
	// Score Against Other Snakes
	for _, s := range nextBoard.Snakes {
		if s.ID == youID {
			continue
		}
		sHead := s.Body[0]
		yHead := you.Body[0]
		dist := abs(sHead.X-yHead.X) + abs(sHead.Y-yHead.Y)
		if dist <= 2 {
			if len(s.Body) >= len(you.Body) {
				if len(safeMoveMap[youID]) == 2 {
					score = score - 5000
				}
				if len(safeMoveMap[youID]) == 3 {
					score = score - 2500
				}
			} else {
				if len(safeMoveMap[youID]) == 2 {
					score = score + 500
				}
				if len(safeMoveMap[youID]) == 3 {
					score = score + 250
				}
			}
		}
		if dist == 0 {
			dist = 1
		}
		snakeScore := int((1.0 / float32(dist)) * 1000)
		if len(s.Body) >= len(you.Body) {
			snakeScore = -snakeScore
		}
		score = score + snakeScore
	}
	return score
}

func abs(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

func min(x int, y int) int {
	if x <= y {
		return x
	}
	return y
}

func max(x int, y int) int {
	if x >= y {
		return x
	}
	return y
}

func buildBoardMap(p Payload) map[string]int {
	boardMap := make(map[string]int)
	for _, s := range p.Board.Snakes {
		for i, c := range s.Body {
			snakeFactor := -10
			if i == 0 {
				continue
			}
			key := keyFromCoord(c)
			if val, ok := boardMap[key]; ok {
				boardMap[key] = val + snakeFactor
			} else {
				boardMap[key] = snakeFactor
			}
		}
	}
	for _, h := range p.Board.Hazards {
		key := keyFromCoord(h)
		if val, ok := boardMap[key]; ok {
			boardMap[key] = val - 10
		} else {
			boardMap[key] = -10
		}
	}
	for _, f := range p.Board.Food {
		key := keyFromCoord(f)
		if val, ok := boardMap[key]; ok {
			boardMap[key] = val + 5
		} else {
			boardMap[key] = 5
		}
		nearbyKeys := splashKeysFromCoord(f, p.Board.Width, p.Board.Height)
		for _, k := range nearbyKeys {
			if val, ok := boardMap[k]; ok {
				boardMap[k] = val + 3
			} else {
				boardMap[k] = 3
			}
		}
	}
	return boardMap
}

func findBestAdjacent(c Coord, boardMap map[string]int) string {
	val := -1000000
	move := "up"
	upVal := boardMap[keyFromCoord(Coord{c.X, c.Y+1})]
	if upVal > val {
		val = upVal
		move = "up"
	}
	downVal := boardMap[keyFromCoord(Coord{c.X, c.Y-1})]
	if downVal > val {
		val = downVal
		move = "down"
	}
	leftVal := boardMap[keyFromCoord(Coord{c.X-1, c.Y})]
	if leftVal > val {
		val = leftVal
		move = "left"
	}
	rightVal := boardMap[keyFromCoord(Coord{c.X+1, c.Y})]
	if rightVal > val {
		val = rightVal
		move = "right"
	}
	return move
}

func keyFromCoord(c Coord) string {
	return strconv.Itoa(int(c.X)) + "-" + strconv.Itoa(int(c.Y))
}

func splashKeysFromCoord(c Coord, w int32, h int32) []string {
	result := make([]string, 0)
	for x := c.X - 2; x <= c.X+2; x++ {
		for y := c.Y - 2; y <= c.Y+2; y++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				result = append(result, keyFromCoord(Coord{x, y}))
			}
		}
	}
	return result
}

func getBoardStateFromBoard(b Board) *rules.BoardState {
	return &rules.BoardState{
		Height: b.Height,
		Width:  b.Width,
		Food:   coordsToPoints(b.Food),
		Snakes: snakesToSnakes(b.Snakes),
	}
}

func coordToPoint(cd Coord) rules.Point {
	return rules.Point{X: cd.X, Y: cd.Y}
}

func coordsToPoints(cdArray []Coord) []rules.Point {
	a := make([]rules.Point, 0)
	for _, cd := range cdArray {
		a = append(a, coordToPoint(cd))
	}
	return a
}

func snakeToSnake(sn Snake) rules.Snake {
	return rules.Snake{
		ID:     sn.Id,
		Body:   coordsToPoints(sn.Body),
		Health: sn.Health,
	}
}

func snakesToSnakes(snArray []Snake) []rules.Snake {
	s := make([]rules.Snake, 0)
	for _, sn := range snArray {
		s = append(s, snakeToSnake(sn))
	}
	return s
}

func isDebug() bool {
        val, ok := os.LookupEnv("ENV")
        if ok && val == "development" {
                return true
        }
        return false
}
