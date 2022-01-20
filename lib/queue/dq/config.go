package dq

import "github.com/gotid/god/lib/store/redis"

type (
	Beanstalk struct {
		Endpoint string
		Tube     string
	}

	Conf struct {
		Beanstalks []Beanstalk
		Redis      redis.Conf
	}
)
