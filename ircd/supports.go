package ircd

import (
	"fmt"
)

var (
	MaxTopicLength = 307
	MaxKickLength  = 307
	MaxAwayLength  = 307
	MaxTargets     = 20
)

func SendSupportLine(cli *Client) {
	conf := cli.Server.Config

	supportLine := fmt.Sprintf(
		"MAXCHANNELS=%d CHANLIMIT=#:%d TOPICLEN=%d KICKLEN=%d MAXTARGETS=%d MAXAWAYLEN=%d STATUSMSG=@%%+ NETWORK=%s PREFIX=(ohv)@%%+ CHANMODES=%s are supported by this server",
		conf.Limits.Channels,
		conf.Limits.Channels,
		MaxTopicLength,
		MaxKickLength,
		MaxTargets,
		MaxAwayLength,
		conf.Network,
		SupportedCModes.String(),
	)

	cli.SendNumeric(RPL_MYINFO, conf.Name, VERSION, SupportedUModes.String(), SupportedCModes.String())
	cli.SendNumeric(RPL_ISUPPORT, supportLine)
}
