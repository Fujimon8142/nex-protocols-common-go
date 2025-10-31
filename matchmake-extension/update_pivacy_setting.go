package matchmake_extension

import (
	"github.com/PretendoNetwork/nex-go/v2"
	"github.com/PretendoNetwork/nex-go/v2/types"
	common_globals "github.com/PretendoNetwork/nex-protocols-common-go/v2/globals"
	"github.com/PretendoNetwork/nex-protocols-common-go/v2/matchmake-extension/database"
	matchmake_extension "github.com/PretendoNetwork/nex-protocols-go/v2/matchmake-extension"
)

func (commonProtocol *CommonProtocol) updatePrivacySetting(err error, packet nex.PacketInterface, callID uint32, onlineStatus types.Bool, participationCommunity types.Bool) (*nex.RMCMessage, *nex.Error) {
	if err != nil {
		common_globals.Logger.Error(err.Error())
		return nil, nex.NewError(nex.ResultCodes.Core.InvalidArgument, "change_error")
	}

	connection := packet.Sender().(*nex.PRUDPConnection)
	endpoint := connection.Endpoint().(*nex.PRUDPEndPoint)

	commonProtocol.manager.Mutex.Lock()
	defer commonProtocol.manager.Mutex.Unlock()

	nexError := database.UpdatePrivacySetting(
		commonProtocol.manager,
		connection.PID(),
		bool(onlineStatus),
		bool(participationCommunity),
	)
	if nexError != nil {
		return nil, nexError
	}

	rmcResponse := nex.NewRMCSuccess(endpoint, nil)
	rmcResponse.ProtocolID = matchmake_extension.ProtocolID
	rmcResponse.MethodID = matchmake_extension.MethodUpdatePrivacySetting
	rmcResponse.CallID = callID

	if commonProtocol.OnAfterUpdatePrivacySetting != nil {
		go commonProtocol.OnAfterUpdatePrivacySetting(packet, onlineStatus, participationCommunity)
	}

	return rmcResponse, nil
}
