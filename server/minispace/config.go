package minispace

type MiniConfig struct {
	EnableDB bool
	EnableAI bool
}

func NewMiniConfig() *MiniConfig {
	return &MiniConfig{
		EnableDB: true,
		EnableAI: true,
	}
}
