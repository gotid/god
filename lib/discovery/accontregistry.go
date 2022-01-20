package discovery

import "git.zc0901.com/go/god/lib/discovery/internal"

func RegisterAccount(endpoints []string, user, pass string) {
	internal.AddAccount(endpoints, user, pass)
}
