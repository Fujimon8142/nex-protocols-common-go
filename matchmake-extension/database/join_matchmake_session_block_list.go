package database

import (
	"github.com/PretendoNetwork/nex-go/v2"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	"github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
)

// JoinMatchmakeSessionBlockList wraps JoinMatchmakeSession with a blocklist check.
// Used by JoinMatchmakeSession and JoinMatchmakeSessionEx handlers.
func JoinMatchmakeSessionBlockList(manager *common_globals.MatchmakingManager, joinedMatchmakeSession types.MatchmakeSession, connection *nex.PRUDPConnection, vacantParticipants uint16, strMessage string, dontCareMyBlockList bool) (uint32, *nex.Error) {
	callerPID := connection.PID()
	gid := uint32(joinedMatchmakeSession.Gathering.ID)

	nexError := ThrowErrorIfBlockListHit(manager, callerPID, gid, dontCareMyBlockList)
	if nexError != nil {
		return 0, nexError
	}

	return JoinMatchmakeSession(manager, joinedMatchmakeSession, connection, vacantParticipants, strMessage)
}
