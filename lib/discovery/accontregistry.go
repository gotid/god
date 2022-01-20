package discovery

import "github.com/gotid/god/lib/discovery/internal"

func RegisterAccount(endpoints []string, user, pass string) {
	internal.AddAccount(endpoints, user, pass)
}
