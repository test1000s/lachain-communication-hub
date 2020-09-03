package peer

import (
	"github.com/juju/loggo"
	"github.com/libp2p/go-libp2p-core/network"
	"io"
	"lachain-communication-hub/communication"
	"time"
)

var handler = loggo.GetLogger("handler")

func incomingConnectionEstablishmentHandler(peer *Peer) func(s network.Stream) {
	return func(s network.Stream) {
		go runHubMsgHandler(peer, s)
	}
}

func runHubMsgHandler(peer *Peer, s network.Stream) {
	handler.Tracef("Running runHubMsgHandler() for peer %p", peer)
	for {
		remotePeerId := s.Conn().RemotePeer()
		connectionExists := peer.IsConnectionWithPeerIdExists(remotePeerId)

		if !connectionExists {
			publicKey, err := peer.GetPeerPublicKeyById(remotePeerId)
			if err != nil {
				s.Close()
				break
			}
			peer.RegisterConnection(publicKey, s)
		}

		msg, err := communication.ReadOnce(s)
		if err != nil {
			if err == io.EOF {
				handler.Errorf("connection reset")
				time.Sleep(2 * time.Second)
				continue
			}
			handler.Errorf("Can't read message. Closing connection")
			handler.Errorf("%s", err)
			s.Close()
			break
		}
		err = processMessage(peer, s, msg)
		if err != nil {
			handler.Errorf("Connection problem")
			s.Close()
			return
		}
	}
}

func processMessage(localPeer *Peer, s network.Stream, msg []byte) error {
	if len(msg) == 0 {
		return nil
	}

	handler.Tracef("Calling grpc message (%p) handler on peer (%p)", localPeer.grpcMsgHandler, localPeer)
	localPeer.grpcMsgHandler(msg)

	handler.Tracef("received msg from peer: %s, message len = %d", s.Conn().RemotePeer(), len(msg))

	switch string(msg) {
	case "ping":
		err := communication.Write(s, []byte("pong"))
		if err != nil {
			return err
		}
		break

		//case "pong":
		//	time.Sleep(2 * time.Second)
		//	_, err := s.Write([]byte("ping"))
		//	if err != nil {
		//		panic(err)
		//	}
		//	break
	}
	return nil
}
