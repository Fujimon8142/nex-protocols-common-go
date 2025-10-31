package database

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	match_making_database "github.com/PretendoNetwork/nex-protocols-common-go/v2/match-making/database"
)

// ThrowErrorIfBlockListHit performs bilateral blocklist check before joining a session.
func ThrowErrorIfBlockListHit(manager *common_globals.MatchmakingManager, callerPID types.PID, gatheringID uint32, dontCareMyBlockList bool) *nex.Error {
	// Get session participants
	_, _, participants, _, nexError := match_making_database.FindGatheringByID(manager, gatheringID)
	if nexError != nil {
		return nexError
	}

	// Get block lists
	var blockList []types.PID
	if !dontCareMyBlockList {
		blockList, nexError = GetBlockList(manager, callerPID)
		if nexError != nil {
			return nexError
		}
	}

	blockedByList, nexError := GetBlockedByList(manager, callerPID)
	if nexError != nil {
		return nexError
	}

	// Check lists against participants
	for _, participantPID := range participants {
		// Check if caller is blocked by a participant
		for _, blockerPID := range blockedByList {
			if types.PID(participantPID) == blockerPID {
				return nex.NewError(nex.ResultCodes.RendezVous.DeniedByParticipants, "change_error")
			}
		}

		// Check if caller has blocked a participant (if not ignored)
		if !dontCareMyBlockList {
			for _, blockedPID := range blockList {
				if types.PID(participantPID) == blockedPID {
					return nex.NewError(nex.ResultCodes.RendezVous.ParticipantInBlackList, "change_error")
				}
			}
		}
	}
	return nil
}
