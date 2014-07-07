package minispace

func Init() error {
	err := ConnectSharedDB("127.0.0.1")
	if err != nil {
		return err
	}

	go CurrentScene().Run()
	return nil
}
