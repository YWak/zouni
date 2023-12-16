package lib

type Number interface {
	int | int32 | int64
}

func Add[V Number](a, b V) V {
	return a + b
}
