package ipfilter

type Service interface {
	IsIPWhitelisted(ip string) bool
	IsIPBlacklisted(ip string) bool
	IsNetworkWhitelisted(network string) (bool, error)
	IsNetworkBlacklisted(network string) (bool, error)
	AddToWhitelist(subnet string) error
	RemoveFromWhitelist(subnet string) (bool, error)
	AddToBlacklist(subnet string) error
	RemoveFromBlacklist(subnet string) (bool, error)
}
