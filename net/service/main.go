package service

import (
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/op/go-logging"
	"github.com/textileio/textile-go/crypto"
	"github.com/textileio/textile-go/pb"
	"github.com/textileio/textile-go/repo"
	inet "gx/ipfs/QmPjvxTpVH8qJyQDnxnsxF9kv9jezKD1kozz1hs3fCGsNh/go-libp2p-net"
	"gx/ipfs/QmTKsRYeY4simJyf37K93juSq75Lo8MVCDJ7owjmf46u8W/go-context/io"
	ggio "gx/ipfs/QmZ4Qi3GaRbjcx28Sme5eMH7RQjGkt8wHxt2a65oLaeFEV/gogo-protobuf/io"
	"gx/ipfs/QmZNkThpqfVXs9GNbexPrfBbXSLNYeKrE7jwFM2oqHbyqN/go-libp2p-protocol"
	"gx/ipfs/QmdVrMn1LhB4ybb8hMVaMLXnA8XRSewMnK6YqXKXoTcRvN/go-libp2p-peer"
	libp2pc "gx/ipfs/Qme1knMqwt1hKZbc1BmQFmnm9f36nyQGwXxPGVpVJ9rMK5/go-libp2p-crypto"
	"gx/ipfs/QmebqVUQQqQFhg74FtQFszUJo22Vpr3e8qBAkvvV4ho9HH/go-ipfs/core"
	"io"
	"math/rand"
	"sync"
	"time"
)

var log = logging.MustGetLogger("net")

// service represents a libp2p service
type Service struct {
	Node      *core.IpfsNode
	Datastore repo.Datastore
	handler   Handler
	sender    map[peer.ID]*sender
	senderMux sync.Mutex
}

// DefaultTimeout is the timeout context for sending / requesting messages
const DefaultTimeout = time.Second * 5

// PeerStatus is the possible results from pinging another peer
type PeerStatus string

const (
	PeerOnline  PeerStatus = "online"
	PeerOffline PeerStatus = "offline"
)

// Handler is used to handle messages for a specific protocol
type Handler interface {
	Protocol() protocol.ID
	Node() *core.IpfsNode
	Datastore() repo.Datastore
	Ping(pid peer.ID) (PeerStatus, error)
	VerifyEnvelope(env *pb.Envelope) error
	Handle(mtype pb.Message_Type) func(peer.ID, *pb.Envelope) (*pb.Envelope, error)
}

// NewService returns a service for the given config
func NewService(
	handler Handler,
	node *core.IpfsNode,
	datastore repo.Datastore,
) *Service {
	service := &Service{
		Node:      node,
		Datastore: datastore,
		handler:   handler,
		sender:    make(map[peer.ID]*sender),
	}
	node.PeerHost.SetStreamHandler(handler.Protocol(), service.handleNewStream)
	log.Infof("registered service: %s", handler.Protocol())
	return service
}

// SendMessage sends a message to a peer
func (s *Service) SendMessage(ctx context.Context, pid peer.ID, pmes *pb.Envelope) error {
	log.Debugf("sending %s message to %s", pmes.Message.Type.String(), pid.Pretty())
	ms, err := s.messageSenderForPeer(pid, s.handler.Protocol())
	if err != nil {
		return err
	}
	if err := ms.SendMessage(ctx, pmes); err != nil {
		return err
	}
	return nil
}

// SendRequest sends a request to a peer
func (s *Service) SendRequest(ctx context.Context, pid peer.ID, pmes *pb.Envelope) (*pb.Envelope, error) {
	log.Debugf("sending %s request to %s", pmes.Message.Type.String(), pid.Pretty())
	ms, err := s.messageSenderForPeer(pid, s.handler.Protocol())
	if err != nil {
		return nil, err
	}
	rpmes, err := ms.SendRequest(ctx, pmes)
	if err != nil {
		log.Debugf("no response from %s", pid.Pretty())
		return nil, err
	}
	if rpmes == nil {
		log.Debugf("no response from %s", pid.Pretty())
		return nil, errors.New("no response from peer")
	}
	log.Debugf("received response from %s", pid.Pretty())
	return rpmes, nil
}

// Ping pings another peer and returns status
func (s *Service) Ping(pid peer.ID) (PeerStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	id := rand.Int31()
	env, err := s.NewEnvelope(pb.Message_PING, nil, &id, false)
	if err != nil {
		return "", err
	}
	if _, err := s.SendRequest(ctx, pid, env); err != nil {
		return PeerOffline, nil
	}
	return PeerOnline, nil
}

// NewEnvelope returns a signed pb message
func (s *Service) NewEnvelope(mtype pb.Message_Type, msg proto.Message, id *int32, response bool) (*pb.Envelope, error) {
	var payload *any.Any
	if msg != nil {
		var err error
		payload, err = ptypes.MarshalAny(msg)
		if err != nil {
			return nil, err
		}
	}
	message := &pb.Message{Type: mtype, Payload: payload}
	if id != nil {
		message.RequestId = *id
	}
	if response {
		message.IsResponse = true
	}
	ser, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	sig, err := s.Node.PrivateKey.Sign(ser)
	if err != nil {
		return nil, err
	}
	pk, err := s.Node.PrivateKey.GetPublic().Bytes()
	if err != nil {
		return nil, err
	}
	env := &pb.Envelope{Message: message, Pk: pk, Sig: sig}
	return env, nil
}

// NewErrorMessage returns a signed pb error message
func (s *Service) NewErrorMessage(code int, msg string) (*pb.Envelope, error) {
	return s.NewEnvelope(pb.Message_ERROR, &pb.Error{Code: uint32(code), Message: msg}, nil, false)
}

// VerifyEnvelope verifies the authenticity of an envelope
func (s *Service) VerifyEnvelope(env *pb.Envelope) error {
	ser, err := proto.Marshal(env.Message)
	if err != nil {
		return err
	}
	pk, err := libp2pc.UnmarshalPublicKey(env.Pk)
	if err != nil {
		return err
	}
	return crypto.Verify(pk, ser, env.Sig)
}

// handleNewStream handles a p2p net stream in the background
func (s *Service) handleNewStream(stream inet.Stream) {
	go s.handleNewMessage(stream)
}

// handleNewMessage handles a p2p net stream
func (s *Service) handleNewMessage(stream inet.Stream) {
	defer stream.Close()

	// setup reader
	ctxReader := ctxio.NewReader(s.Node.Context(), stream)
	reader := ggio.NewDelimitedReader(ctxReader, inet.MessageSizeMax)

	// get sender
	rpid := stream.Conn().RemotePeer()
	ms, err := s.messageSenderForPeer(rpid, s.handler.Protocol())
	if err != nil {
		log.Error("error getting message sender")
		return
	}

	// start listening for messages from this sender
	for {
		select {
		// end loop on context close
		case <-s.Node.Context().Done():
			return
		default:
		}

		// receive msg
		env := new(pb.Envelope)
		if err := reader.ReadMsg(env); err != nil {
			stream.Reset()
			if err == io.EOF {
				log.Debugf("disconnected from peer %s", rpid.Pretty())
			}
			return
		}

		// check signature
		if err := s.VerifyEnvelope(env); err != nil {
			log.Warningf("error verifying message: %s", err)
			return
		}

		// check if the message is a response
		if env.Message.IsResponse {
			ms.requestMux.Lock()
			ch, ok := ms.requests[env.Message.RequestId]
			if ok {
				// this is a request response
				select {
				case ch <- env:
					// message returned to requester
				case <-time.After(time.Second):
					// in case ch is closed on the other end - the lock should prevent this happening
					log.Debug("request id was not removed from map on timeout")
				}
				close(ch)
				delete(ms.requests, env.Message.RequestId)
			} else {
				log.Debug("unknown request id: requesting function may have timed out")
			}
			ms.requestMux.Unlock()
			stream.Reset()
			return
		}

		// try a generic handler for this msg type
		handler := s.handleGeneric(env.Message.Type)
		if handler == nil {
			// get service specific handler for this msg type
			handler := s.handler.Handle(env.Message.Type)
			if handler == nil {
				stream.Reset()
				log.Debug("got back nil handler")
				return
			}
		}

		// dispatch handler
		renv, err := handler(rpid, env)
		if err != nil {
			log.Errorf("%s handle message error: %s", env.Message.Type.String(), err)
		}

		// if nil response, return it before serializing
		if renv == nil {
			continue
		}

		// send out response msg
		if err := ms.SendMessage(s.Node.Context(), renv); err != nil {
			stream.Reset()
			log.Errorf("send response error: %s", err)
			return
		}
	}
}

// handleGeneric provides service level handlers for common message types
func (s *Service) handleGeneric(mtype pb.Message_Type) func(peer.ID, *pb.Envelope) (*pb.Envelope, error) {
	switch mtype {
	case pb.Message_PING:
		return s.handlePing
	case pb.Message_ERROR:
		return s.handleError
	default:
		return nil
	}
}

// handlePing receives a PING message
func (s *Service) handlePing(pid peer.ID, env *pb.Envelope) (*pb.Envelope, error) {
	log.Debugf("received PING message from %s", pid.Pretty())
	return s.NewEnvelope(pb.Message_PONG, nil, &env.Message.RequestId, true)
}

// handleError receives an ERROR message
func (s *Service) handleError(pid peer.ID, env *pb.Envelope) (*pb.Envelope, error) {
	log.Debugf("received ERROR message from %s", pid.Pretty())
	if env.Message.Payload == nil {
		return nil, errors.New("payload is nil")
	}
	errorMessage := new(pb.Error)
	err := ptypes.UnmarshalAny(env.Message.Payload, errorMessage)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
