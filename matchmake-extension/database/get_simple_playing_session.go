package database

import (
	"database/sql"

	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	match_making_types "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
	pqextended "github.com/PretendoNetwork/pq-extended"
)

// GetSimplePlayingSession returns the simple playing sessions of the given PIDs
func GetSimplePlayingSession(manager *common_globals.MatchmakingManager, callerPID types.PID, listPID []types.PID, friendList []uint32) ([]match_making_types.SimplePlayingSession, *nex.Error) {
	simplePlayingSessions := make([]match_making_types.SimplePlayingSession, 0)
	for _, pid := range listPID {
		simplePlayingSession := match_making_types.NewSimplePlayingSession()

		err := manager.Database.QueryRow(`SELECT
		g.id,
		ms.attribs[1],
		ms.game_mode
		FROM matchmaking.gatherings AS g
		INNER JOIN matchmaking.matchmake_sessions AS ms ON ms.id = g.id
		WHERE
		g.registered=true AND
		g.type='MatchmakeSession' AND
		$1=ANY(g.participants) AND
		NOT EXISTS (
			SELECT 1
			FROM matchmaking.block_lists bl
			WHERE bl.user_pid = $1 AND bl.blocked_pid = $2
		)
		AND (
			$1 = ANY($3) -- Target user is a friend
			OR
			-- Target user is not a friend, check privacy settings
			-- COALESCE is used to default to 'true' (visible) if no row exists
			COALESCE(
				(SELECT ps.online_status
				FROM matchmaking.privacy_settings ps
				WHERE ps.user_pid = $1),
				true
			)
		)`, pid, callerPID, pqextended.Array(friendList)).Scan(
		&simplePlayingSession.GatheringID,
		&simplePlayingSession.Attribute0,
		&simplePlayingSession.GameMode)
		if err != nil {
			if err != sql.ErrNoRows {
				common_globals.Logger.Critical(err.Error())
			}
			continue
		}

		simplePlayingSession.PrincipalID = pid

		simplePlayingSessions = append(simplePlayingSessions, simplePlayingSession)
	}

	return simplePlayingSessions, nil
}
