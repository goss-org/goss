package system

type Service interface {
	Service() string
	Enabled() (interface{}, error)
	Running() (interface{}, error)
}
