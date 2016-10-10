package steam

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func mustn(_ int, err error) {
	if err != nil {
		panic(err)
	}
}
