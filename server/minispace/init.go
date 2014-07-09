package minispace

var miniConfig *MiniConfig

func Init(c *MiniConfig) error {
	miniConfig = c

	if miniConfig.EnableDB {
		err := ConnectSharedDB("127.0.0.1")
		if err != nil {
			return err
		}
	}

	go CurrentScene().Run()
	return nil
}
