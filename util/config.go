package util

type Config struct {
	IgnoreList        []string
	Timeout           int
	AllowInsecure     bool
	NoFollowRedirects bool
	Server            string
}
