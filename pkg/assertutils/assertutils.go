package assertutils

func PanicIfNil(o any) {
	if o == nil {
		panic("Nil")
	}
}
