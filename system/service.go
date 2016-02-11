package system

type Service interface {
	Service() string
	Exists() (bool, error)
	Enabled() (bool, error)
	Running() (bool, error)
}
