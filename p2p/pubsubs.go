package p2p

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/pkg/errors"
)

func NewPubSubs(handle HandleSubscriptionMessage, bootNodeMultiaddr string, iotexChainID int) (*PubSubs, error) {
	ctx := context.Background()
	h, err := libp2p.New(libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"), libp2p.Muxer("/yamux/2.0.0", yamux.DefaultTransport))
	if err != nil {
		return nil, errors.Wrap(err, "new libp2p host failed")
	}

	ps, err := pubsub.NewGossipSub(ctx, h, pubsub.WithMaxMessageSize(2*pubsub.DefaultMaxMessageSize))
	if err != nil {
		return nil, errors.Wrap(err, "new gossip subscription failed")
	}
	if err := discoverPeers(ctx, h, bootNodeMultiaddr, iotexChainID); err != nil {
		return nil, err
	}

	return &PubSubs{
		ps:      ps,
		pubSubs: make(map[uint64]*projectPubSub),
		selfID:  h.ID(),
		handle:  handle,
	}, nil
}

type PubSubs struct {
	mux     sync.RWMutex
	pubSubs map[uint64]*projectPubSub
	ps      *pubsub.PubSub
	selfID  peer.ID
	handle  HandleSubscriptionMessage
}

func (p *PubSubs) Add(projectID uint64) error {
	p.mux.Lock()
	defer p.mux.Unlock()

	if _, ok := p.pubSubs[projectID]; ok {
		return nil
	}

	nps, err := newProjectPubSub(projectID, p.ps, p.handle, p.selfID)
	if err != nil {
		return err
	}

	p.pubSubs[projectID] = nps
	return nil
}

func (p *PubSubs) Delete(projectID uint64) {
	p.mux.Lock()
	defer p.mux.Unlock()

	pubSub, ok := p.pubSubs[projectID]
	if !ok {
		return
	}
	pubSub.release()
	delete(p.pubSubs, projectID)
}

func (p *PubSubs) get(projectID uint64) (*projectPubSub, bool) {
	p.mux.RLock()
	defer p.mux.RUnlock()

	s, ok := p.pubSubs[projectID]
	return s, ok
}

func (p *PubSubs) Publish(projectID uint64, data *Data) error {
	d, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "failed to marshal p2p data")
	}

	s, ok := p.get(projectID)
	if !ok {
		return errors.Errorf("project %v topic not exist", projectID)
	}
	if err := s.topic.Publish(context.Background(), d); err != nil {
		return errors.Wrap(err, "failed to publish data to p2p network")
	}
	return nil
}
