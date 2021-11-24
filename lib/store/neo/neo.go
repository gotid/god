package neo

func NewNeo(target, username, password, realm string) Driver {
	return NewDriver(target, username, password, realm)
}
