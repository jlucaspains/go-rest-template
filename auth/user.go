package auth

type User struct {
	ID     string
	Name   string
	Email  string
	Claims map[string]string
}
