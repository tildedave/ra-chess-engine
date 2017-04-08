package main

// http://stackoverflow.com/questions/28541609/looking-for-reasonable-stack-implementation-in-golang
// rather than pulling in a dependency

type byteStack struct {
	arr []byte
}

func (s *byteStack) Push(v byte) {
	s.arr = append(s.arr, v)
}

func (s *byteStack) Pop() byte {
	l := len(s.arr)
	i := s.arr[l-1]
	s.arr = s.arr[:l-1]

	return i
}
