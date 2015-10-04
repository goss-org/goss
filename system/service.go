package system

type Service interface {
	Service() string
	Exists() (interface{}, error)
	Enabled() (interface{}, error)
	Running() (interface{}, error)
}
