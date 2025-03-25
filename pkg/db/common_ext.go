package db

type Ludomans []Ludoman

func (l Ludomans) IDs() []int {
	var res []int
	for _, ludoman := range l {
		res = append(res, ludoman.ID)
	}
	return res
}
