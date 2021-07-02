package snake

import (
	"github.com/BattlesnakeOfficial/rules"
	"strconv"
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

func Move(p Payload) string {
	var ruleset rules.Ruleset
	standard := rules.StandardRuleset{
		FoodSpawnChance: 15,
		MinimumFood:     1,
	}
	ruleset = &standard
	boardState := getBoardStateFromBoard(p.Board)
	boardMap := buildBoardMap(p)
	return tryMoves(p.You.Id, ruleset, boardState, boardMap)
}

func buildBoardMap(p Payload) map[string]int {
	boardMap := make(map[string]int)
	for _, s := range p.Board.Snakes {
		for i, c := range s.Body {
			snakeFactor := -10
			if i == 0 && p.You.Length > s.Length {
				snakeFactor = int(s.Length)
			}
			key := keyFromCoord(c)
			if val, ok := boardMap[key]; ok {
				boardMap[key] = val + snakeFactor
			} else {
				boardMap[key] = snakeFactor
			}
			if i == 0 && s.Id != p.You.Id {
				nearbyKeys := splashKeysFromCoord(c, p.Board.Width, p.Board.Height)
				for _, k := range nearbyKeys {
					if val, ok := boardMap[k]; ok {
						boardMap[k] = val + (snakeFactor/2)
					} else {
						boardMap[k] = snakeFactor / 2
					}
				}
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

func keyFromCoord(c Coord) string {
	return strconv.Itoa(int(c.X)) + "-" + strconv.Itoa(int(c.Y))
}

func keyFromPoint(p rules.Point) string {
	return strconv.Itoa(int(p.X)) + "-" + strconv.Itoa(int(p.Y))
}

func splashKeysFromCoord(c Coord, w int32, h int32) []string {
	result := make([]string, 0)
	for x := c.X - 2; x <= c.X + 2; x++ {
		for y := c.Y - 2; y <= c.Y + 2; y++ {
			if x >= 0 && x < w && y >= 0 && y < h {
				result = append(result, keyFromCoord(Coord{x,y}))
			}
		}
	}
	return result
}

func tryMoves(you string, r rules.Ruleset, b *rules.BoardState, boardMap map[string]int) string {
	possibleMoves := [4]string{"up", "down", "left", "right"}
	arr := make([]string, len(b.Snakes))
	gradeMap := make(map[string]map[string]int)
	generateGrades(arr, 0, r, b, gradeMap, boardMap)
	youIndex := 0
	for i, s := range b.Snakes {
		if you == s.ID {
			youIndex = i
		}
	}
	scoreArray := [4]int{0,0,0,0}
	countArray := [4]int{0,0,0,0}
	for moveKey, scoreMap := range gradeMap {
		yourMove := moveKey[youIndex] - byte('0')
		scoreArray[yourMove] = scoreArray[yourMove] + scoreMap[you]
		countArray[yourMove] = countArray[yourMove] + 1
	}
	highScore := 0
	move := "up"
	for i, m := range possibleMoves {
		avg := scoreArray[i] / countArray[i]
		if avg > highScore {
			highScore = avg
			move = m
		}
	}
	// TODO: Iterate the board a few more times. Maybe take out a bunch of the hokey shit and just assume everyone is trying to do the same thing.
	return move
}

func generateGrades(arr []string, curPos int, r rules.Ruleset,
	b *rules.BoardState, gradeMap map[string]map[string]int,
	boardMap map[string]int) {
	possibleMoves := [4]string{"up", "down", "left", "right"}
	if curPos == len(arr) {
		snakeMoves := make([]rules.SnakeMove, 0)
		moveKey := ""
		for i := 0; i < len(b.Snakes); i++ {
			snakeMoves = append(snakeMoves, rules.SnakeMove{
				ID:   b.Snakes[i].ID,
				Move: arr[i],
			})
			for j, d := range possibleMoves {
				if d == arr[i] {
					moveKey = moveKey + strconv.Itoa(j)
				}
			}
		}
		gradeMap[moveKey] = make(map[string]int)
		result, _ := r.CreateNextBoardState(b, snakeMoves)
		for _, s := range result.Snakes {
			key := keyFromPoint(s.Body[0])
                        targetScore := boardMap[key]
			score := len(s.Body) + targetScore
			if len(s.EliminatedCause) > 0 {
				score = 0
			}
			gradeMap[moveKey][s.ID] = score
		}
		return
	}
	for _, val := range possibleMoves {
		arr[curPos] = val
		generateGrades(arr, curPos+1, r, b, gradeMap, boardMap)
	}
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
