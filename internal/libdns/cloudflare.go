//go:build libdns_cloudflare || !libdns_separate
// +build libdns_cloudflare !libdns_separate

package libdns

import (
	"github.com/libdns/cloudflare"
	"github.com/mail-chat-chain/sirrmeshd/framework/config"
	"github.com/mail-chat-chain/sirrmeshd/framework/module"
)

func init() {
	module.Register("libdns.cloudflare", func(modName, instName string, _, _ []string) (module.Module, error) {
		p := cloudflare.Provider{}
		return &ProviderModule{
			RecordDeleter:  &p,
			RecordAppender: &p,
			setConfig: func(c *config.Map) {
				c.String("api_token", false, false, "", &p.APIToken)
			},
			instName: instName,
			modName:  modName,
		}, nil
	})
}
