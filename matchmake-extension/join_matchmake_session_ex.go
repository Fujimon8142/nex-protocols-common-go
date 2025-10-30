package matchmake_extension

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	match_making_database "github.com/PretendoNetwork/nex-protocols-common-go/v2/match-making/database"
	"github.com/PretendoNetwork/nex-protocols-common-go/v2/matchmake-extension/database"
	matchmake_extension "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"
)

func (commonProtocol *CommonProtocol) joinMatchmakeSessionEx(err error, packet nex.PacketInterface, callID uint32, gid types.UInt32, strMessage types.String, dontCareMyBlockList types.Bool, participationCount types.UInt16) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		common_globals.Logger.Error(err.Error())
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}

	if len(strMessage) > 256 {
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}

	commonProtocol.manager.Mutex.Lock()

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)
	server := endpoint.Server
	callerPID := connection.PID() // Get caller PID

	joinedMatchmakeSession, _, nexError := database.GetMatchmakeSessionByID(commonProtocol.manager, endpoint, uint32(gid))
	if nexError != nil {
		common_globals.Logger.Error(nexError.Error())
		commonProtocol.manager.Mutex.Unlock()
		return nil, nexError
	}

	// --- BEGIN BLOCKLIST CHECK ---
	// Get session participants
	_, _, participants, _, nexError := match_making_database.FindGatheringByID(commonProtocol.manager, uint32(gid))
	if nexError != nil {
		commonProtocol.manager.Mutex.Unlock()
		return nil, nexError
	}

	// Get block lists
	blockedByList, nexError := database.GetBlockedByList(commonProtocol.manager, callerPID)
	if nexError != nil {
		commonProtocol.manager.Mutex.Unlock()
		return nil, nexError
	}

	// Check if caller is blocked by a participant
	for _, participantPID := range participants {
		for _, blockerPID := range blockedByList {
			if types.PID(participantPID) == blockerPID {
				commonProtocol.manager.Mutex.Unlock()
				return nil, nex.NewError(nex.ResultCodes.RendezVous.DeniedByParticipants, "change_error") // RendezVous::DeniedByParticipants
			}
		}
	}

	// Check if caller has blocked a participant (if dontCareMyBlockList is false)
	if !bool(dontCareMyBlockList) {
		blockList, nexError := database.GetBlockList(commonProtocol.manager, callerPID)
		if nexError != nil {
			commonProtocol.manager.Mutex.Unlock()
			return nil, nexError
		}

		for _, participantPID := range participants {
			for _, blockedPID := range blockList {
				if types.PID(participantPID) == blockedPID {
					commonProtocol.manager.Mutex.Unlock()
					return nil, nex.NewError(nex.ResultCodes.RendezVous.ParticipantInBlackList, "change_error") // RendezVous::ParticipantInBlackList
				}
			}
		}
	}
	// --- END BLOCKLIST CHECK ---

	// TODO - Is this the correct error code?
	if joinedMatchmakeSession.UserPasswordEnabled || joinedMatchmakeSession.SystemPasswordEnabled {
		commonProtocol.manager.Mutex.Unlock()
		return nil, nex.NewError(nex.ResultCodes.RendezVous.PermissionDenied, "change_error")
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

	_, nexError = database.JoinMatchmakeSession(commonProtocol.manager, joinedMatchmakeSession, connection, uint16(participationCount), string(strMessage))
	if nexError != nil {
		common_globals.Logger.Error(nexError.Error())
		commonProtocol.manager.Mutex.Unlock()
		return nil, nexError
	}

	commonProtocol.manager.Mutex.Unlock()

	rmcResponseStream := nex.NewByteStreamOut(endpoint.LibraryVersions(), endpoint.ByteStreamSettings())

	if server.LibraryVersions.MatchMaking.GreaterOrEqual("3.0.0") {
		joinedMatchmakeSession.SessionKey.WriteTo(rmcResponseStream)
	}

	rmcResponseBody := rmcResponseStream.Bytes()

	rmcResponse := nex.NewRMCSuccess(endpoint, rmcResponseBody)
	rmcResponse.ProtocolID = matchmake_extension.ProtocolID
	rmcResponse.MethodID = matchmake_extension.MethodJoinMatchmakeSessionEx
	rmcResponse.CallID = callID

	if commonProtocol.OnAfterJoinMatchmakeSessionEx != nil {
		go commonProtocol.OnAfterJoinMatchmakeSessionEx(packet, gid, strMessage, dontCareMyBlockList, participationCount)
	}

	return rmcResponse, nil
}
