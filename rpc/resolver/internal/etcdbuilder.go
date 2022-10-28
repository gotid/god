package internal

type etcdBuilder struct {
	discovBuilder
}

func (e *etcdBuilder) Scheme() string {
	return EtcdSchema
}
