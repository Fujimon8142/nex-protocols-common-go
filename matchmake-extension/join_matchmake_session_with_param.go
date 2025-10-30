package matchmake_extension

import (
	"github.com/PretendoNetwork/nex-go/v2"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	match_making_database "github.com/PretendoNetwork/nex-protocols-common-go/v2/match-making/database"
	"github.com/PretendoNetwork/nex-protocols-common-go/v2/matchmake-extension/database"
	"github.com/PretendoNetwork/nex-protocols-go/v2/match-making/constants"
	match_making_types "github.com/PretendoNetwork/nex-protocols-go/v2/match-making/types"
	matchmake_extension "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"
	"github.com/PretendoNetwork/nex-go/v2/types"
)

func (commonProtocol *CommonProtocol) joinMatchmakeSessionWithParam(err error, packet nex.PacketInterface, callID uint32, joinMatchmakeSessionParam match_making_types.JoinMatchmakeSessionParam) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		common_globals.Logger.Error(err.Error())
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}

	if len(joinMatchmakeSessionParam.JoinMessage) > 256 {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}

	commonProtocol.manager.Mutex.Lock()

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)
	callerPID := connection.PID() // Get caller PID

	if joinMatchmakeSessionParam.GIDForParticipationCheck != 0 {
		// * Check that all new participants are participating in the specified gathering ID
		nexError := database.CheckGatheringForParticipation(commonProtocol.manager, uint32(joinMatchmakeSessionParam.GIDForParticipationCheck), append(joinMatchmakeSessionParam.AdditionalParticipants, connection.PID()))
		if nexError != nil {
			commonProtocol.manager.Mutex.Unlock()
			return nil, nexError
		}
	}

	joinedMatchmakeSession, systemPassword, nexError := database.GetMatchmakeSessionByID(commonProtocol.manager, endpoint, uint32(joinMatchmakeSessionParam.GID))
	if nexError != nil {
		common_globals.Logger.Error(nexError.Error())
		commonProtocol.manager.Mutex.Unlock()
		return nil, nexError
	}

	// --- BEGIN BLOCKLIST CHECK ---
	// Get session participants
	_, _, participants, _, nexError := match_making_database.FindGatheringByID(commonProtocol.manager, uint32(joinMatchmakeSessionParam.GID))
	if nexError != nil {
		commonProtocol.manager.Mutex.Unlock()
		return nil, nexError
	}

	// Get block lists
	blockList, nexError := database.GetBlockList(commonProtocol.manager, callerPID)
	if nexError != nil {
		commonProtocol.manager.Mutex.Unlock()
		return nil, nexError
	}

	blockedByList, nexError := database.GetBlockedByList(commonProtocol.manager, callerPID)
	if nexError != nil {
		commonProtocol.manager.Mutex.Unlock()
		return nil, nexError
	}

	// Check lists against participants
	// Note: This param struct doesn't have dontCareMyBlockList, so we check both lists.
	for _, participantPID := range participants {
		// Check if caller is blocked by a participant
		for _, blockerPID := range blockedByList {
			if types.PID(participantPID) == blockerPID {
				commonProtocol.manager.Mutex.Unlock()
				return nil, nex.NewError(nex.ResultCodes.RendezVous.DeniedByParticipants, "change_error") // RendezVous::DeniedByParticipants
			}
		}

		// Check if caller has blocked a participant
		for _, blockedPID := range blockList {
			if types.PID(participantPID) == blockedPID {
				commonProtocol.manager.Mutex.Unlock()
				return nil, nex.NewError(nex.ResultCodes.RendezVous.ParticipantInBlackList, "change_error") // RendezVous::ParticipantInBlackList
			}
		}
	}
	// --- END BLOCKLIST CHECK ---

	// TODO - Are these the correct error codes?
	if bool(joinedMatchmakeSession.UserPasswordEnabled) && !joinMatchmakeSessionParam.StrUserPassword.Equals(joinedMatchmakeSession.UserPassword) {
		commonProtocol.manager.Mutex.Unlock()
		return nil, nex.NewError(nex.ResultCodes.RendezVous.InvalidPassword, "change_error")
	}

	if bool(joinedMatchmakeSession.SystemPasswordEnabled) && string(joinMatchmakeSessionParam.StrSystemPassword) != systemPassword {
		commonProtocol.manager.Mutex.Unlock()
		return nil, nex.NewError(nex.ResultCodes.RendezVous.InvalidPassword, "change_error")
	}

	// * Allow game servers to do their own permissions checks
	if commonProtocol.CanJoinMatchmakeSession != nil {
		nexError = commonProtocol.CanJoinMatchmakeSession(commonProtocol.manager, connection.PID(), joinedMatchmakeSession)
	} else {
		nexError = common_globals.CanJoinMatchmakeSession(commonProtocol.manager, connection.PID(), joinedMatchmakeSession)
	}
	if nexError != nil {
		commonProtocol.manager.Mutex.Unlock()
		return nil, nexError
	}

	_, nexError = database.JoinMatchmakeSessionWithParticipants(commonProtocol.manager, joinedMatchmakeSession, connection, joinMatchmakeSessionParam.AdditionalParticipants, string(joinMatchmakeSessionParam.JoinMessage), constants.JoinMatchmakeSessionBehavior(joinMatchmakeSessionParam.JoinMatchmakeSessionBehavior))
	if nexError != nil {
		common_globals.Logger.Error(nexError.Error())
		commonProtocol.manager.Mutex.Unlock()
		return nil, nexError
	}

	commonProtocol.manager.Mutex.Unlock()

	rmcResponseStream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())

	joinedMatchmakeSession.WriteTo(rmcResponseStream)

	rmcResponseBody := rmcResponseStream.Bytes()

	rmcResponse := nex.NewRMCSuccess(endpoint, rmcResponseBody)
	rmcResponse.ProtocolID = matchmake_extension.ProtocolID
	rmcResponse.MethodID = matchmake_extension.MethodJoinMatchmakeSessionWithParam
	rmcResponse.CallID = callID

	if commonProtocol.OnAfterJoinMatchmakeSessionWithParam != nil {
		go commonProtocol.OnAfterJoinMatchmakeSessionWithParam(packet, joinMatchmakeSessionParam)
	}

	return rmcResponse, nil
}
