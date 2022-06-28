package mkzip

func perr(err error) {
	if err != nil {
		panic(err)
	}
}
