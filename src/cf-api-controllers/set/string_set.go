package set

type StringSet map[string]struct{}

func NewStringSet(s []string) StringSet {
	res := StringSet{}
	for _, v := range s {
		res[v] = struct{}{}
	}
	return res
}

func (s StringSet) Difference(right StringSet) StringSet {
	res := StringSet{}
	for k := range s {
		if _, present := right[k]; !present {
			res[k] = struct{}{}
		}
	}
	return res
}

func (s StringSet) ToSlice() []string {
	res := make([]string, 0)
	for k := range s {
		res = append(res, k)
	}
	return res
}
