package http

import (
	"minerva.devops.letv.com/scloud/stargazer-base-lib/config"
)

func init() {
	config.AddHook(
		func() {
			LoadErrorCode()
		})
}
