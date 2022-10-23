package internal

type etcdBuilder struct {
	discovBuilder
}

func (e *etcdBuilder) Schema() string {
	return EtcdSchema
}
