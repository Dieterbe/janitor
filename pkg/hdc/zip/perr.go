package zip

func perr(err error) {
	if err != nil {
		panic(err)
	}
}
