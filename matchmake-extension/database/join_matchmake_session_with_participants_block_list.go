package database

import (
	"github.com/PretendoNetwork/nex-go/v2"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	"github.com/PretendoNetwork/nex-protocols-go/v2/match-making/constants"
	"github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
)

// JoinMatchmakeSessionWithParticipantsBlockList wraps JoinMatchmakeSessionWithParticipants with a blocklist check.
// Used by JoinMatchmakeSessionWithParam handler.
func JoinMatchmakeSessionWithParticipantsBlockList(manager *common_globals.MatchmakingManager, joinedMatchmakeSession types.MatchmakeSession, connection *nex.PRUDPConnection, joinMatchmakeSessionParam types.JoinMatchmakeSessionParam) (uint32, *nex.Error) {
	callerPID := connection.PID()
	gid := uint32(joinedMatchmakeSession.Gathering.ID)
	dontCareMyBlockList := false // JoinMatchmakeSessionWithParam doesn't have this option

	nexError := ThrowErrorIfBlockListHit(manager, callerPID, gid, dontCareMyBlockList)
	if nexError != nil {
		return 0, nexError
	}

	return JoinMatchmakeSessionWithParticipants(
		manager,
		joinedMatchmakeSession,
		connection,
		joinMatchmakeSessionParam.AdditionalParticipants,
		string(joinMatchmakeSessionParam.JoinMessage),
		constants.JoinMatchmakeSessionBehavior(joinMatchmakeSessionParam.JoinMatchmakeSessionBehavior),
	)
}
