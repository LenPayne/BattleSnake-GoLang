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
	Latency int32   `json:"latency"`
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
	return tryMoves(p.You.Id, ruleset, boardState)
}

/*
	For your snake, try u/d/l/r
	For each other snake also try u/d/l/r
	This is 4^8 options max (65k)
	The best option seems to be:
	 - Simulate all the possibilities and grade them for each snake
         - Assume everyone is trying to maximize their own results
	 - Assume everyone is trying to minimize other snake results
*/
func tryMoves(you string, r rules.Ruleset, b *rules.BoardState) string {
	possibleMoves := [4]string{"up", "down", "left", "right"}
	arr := make([]string, len(b.Snakes))
	gradeMap := make(map[string]map[string]int)
	generateGrades(arr, 0, r, b, gradeMap)
	maxMoveScore := make(map[string]int)
	maxMoveMove := make(map[string]string)
	for _, s := range b.Snakes {
		maxMoveScore[s.ID] = -1
		maxMoveMove[s.ID] = "up"
	}
	for moveKey, scoreMap := range gradeMap {
		for i, s := range b.Snakes {
			if scoreMap[s.ID] > maxMoveScore[s.ID] {
				maxMoveScore[s.ID] = scoreMap[s.ID]
				snakeIndex := moveKey[i] - byte('0')
				maxMoveMove[s.ID] = possibleMoves[snakeIndex]
			}
		}
	}
	return maxMoveMove[you]
}

func generateGrades(arr []string, curPos int, r rules.Ruleset,
	b *rules.BoardState, gradeMap map[string]map[string]int) {
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
			score := len(s.Body)
			if len(s.EliminatedCause) > 0 {
				score = 0
			}
			gradeMap[moveKey][s.ID] = score
		}
		return
	}
	for _, val := range possibleMoves {
		arr[curPos] = val
		generateGrades(arr, curPos+1, r, b, gradeMap)
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
