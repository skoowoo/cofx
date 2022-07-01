package cofunc

type _Var struct {
	v         string
	child     []*_Var
	cacheable bool
}
