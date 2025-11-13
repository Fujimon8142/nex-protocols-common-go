package database

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	pqextended "github.com/PretendoNetwork/pq-extended"
)

// RemoveFromBlockList removes a list of PIDs from a user's blocklist
func RemoveFromBlockList(manager *common_globals.MatchmakingManager, userPID types.PID, pidsToUnblock types.List[types.PID]) *nex.Error {

	pids := make([]uint64, len(pidsToUnblock))
	for i, pid := range pidsToUnblock {
		pids[i] = uint64(pid)
	}

	// --- DEBUGログ追加 (1/2) ---
	common_globals.Logger.Infof("[DB] RemoveFromBlockList: UserPID %d, PIDsToUnblock %v", userPID, pids)
	// --- DEBUGログ追加ここまで ---

	result, err := manager.Database.Exec(`
		DELETE FROM matchmaking.block_lists 
		WHERE user_pid = $1 AND blocked_pid = ANY($2)
	`, uint64(userPID), pqextended.Array(pids))

	if err != nil {
		// --- DEBUGログ追加 (2/2) ---
		common_globals.Logger.Errorf("[DB] RemoveFromBlockList Exec error for UserPID %d: %s", userPID, err.Error())
		// --- DEBUGログ追加ここまで ---
		return nex.NewError(nex.ResultCodes.Core.Unknown, err.Error())
	}

	// --- DEBUGログ追加 (2/2) ---
	rowsAffected, _ := result.RowsAffected()
	common_globals.Logger.Infof("[DB] RemoveFromBlockList: UserPID %d successfully deleted %d rows.", userPID, rowsAffected)
	// --- DEBUGログ追加ここまで ---

	return nil
}
