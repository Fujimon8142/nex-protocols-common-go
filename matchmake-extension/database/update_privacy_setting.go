package database

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
)

// UpdatePrivacySetting updates or inserts privacy settings for a user.
func UpdatePrivacySetting(manager *common_globals.MatchmakingManager, userPID types.PID, onlineStatus bool, participationCommunity bool) *nex.Error {
	_, err := manager.Database.Exec(`
		INSERT INTO matchmaking.privacy_settings (user_pid, online_status, participation_community)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_pid)
		DO UPDATE SET
			online_status = $2,
			participation_community = $3
	`, uint64(userPID), onlineStatus, participationCommunity)

	if err != nil {
		return nex.NewError(nex.ResultCodes.Core.Unknown, err.Error())
	}

	return nil
}
