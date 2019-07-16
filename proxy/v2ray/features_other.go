// +build !ios,!android

package v2ray

import (
	_ "v2ray.com/core/app/commander"
	_ "v2ray.com/core/app/log/command"
	_ "v2ray.com/core/app/proxyman/command"
	_ "v2ray.com/core/app/stats/command"

	_ "v2ray.com/core/app/reverse"

	_ "v2ray.com/core/transport/internet/domainsocket"
)
