package matchmake_extension

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	"github.com/PretendoNetwork/nex-protocols-common-go/v2/matchmake-extension/database"
	matchmake_extension "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"
)

func (commonProtocol *CommonProtocol) removeFromBlockList(err error, packet nex.PacketInterface, callID uint32, lstPrincipalID types.List[types.PID]) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		common_globals.Logger.Error(err.Error())
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)

	// --- DEBUGログ追加 (1/2) ---
	common_globals.Logger.Infof("RemoveFromBlockList called by PID: %d", connection.PID())
	common_globals.Logger.Infof("Attempting to unblock PIDs: %v", lstPrincipalID)
	// --- DEBUGログ追加ここまで ---

	commonProtocol.manager.Mutex.Lock()
	defer commonProtocol.manager.Mutex.Unlock()

	nexError := database.RemoveFromBlockList(commonProtocol.manager, connection.PID(), lstPrincipalID)
	if nexError != nil {
		// --- DEBUGログ追加 (2/2) ---
		common_globals.Logger.Errorf("RemoveFromBlockList failed for PID %d: %s", connection.PID(), nexError.Error())
		// --- DEBUGログ追加ここまで ---
		return nil, nexError
	}

	// --- DEBUGログ追加 (2/2) ---
	common_globals.Logger.Infof("RemoveFromBlockList successful for PID %d.", connection.PID())
	// --- DEBUGログ追加ここまで ---

	rmcResponse := nex.NewRMCSuccess(endpoint, nil)
	rmcResponse.ProtocolID = matchmake_extension.ProtocolID
	rmcResponse.MethodID = matchmake_extension.MethodRemoveFromBlockList
	rmcResponse.CallID = callID

	if commonProtocol.OnAfterRemoveFromBlockList != nil {
		go commonProtocol.OnAfterRemoveFromBlockList(packet, lstPrincipalID)
	}

	return rmcResponse, nil
}
