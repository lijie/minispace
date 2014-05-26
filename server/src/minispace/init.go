package minispace

func Init() error {
	go CurrentScene().Run()
	return nil
}
