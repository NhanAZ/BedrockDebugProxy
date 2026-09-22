package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ClientCacheBlobStatus is part of the blob cache protocol. It is sent by the client to let the server know
// what blobs it needs and which blobs it already has, in an ACK type system.
type ClientCacheBlobStatus struct {
	// MissHashes is a list of blob hashes that the client does not have a blob available for. The server
	// should send the blobs matching these hashes as soon as possible.
	MissHashes []uint64
	// HitHashes is a list of blob hashes that the client has a blob available for. The blobs hashes here mean
	// that the client already has them: The server does not need to send the blobs anymore.
	HitHashes []uint64
}

// A featured-experience client can report more than the generic protocol slice
// bound when its cache spans many chunks. Keep the field-specific bound finite
// while allowing the observed 1,302-hash status packet.
const maxClientCacheBlobHashes = 16 * 1024

// ID ...
func (pk *ClientCacheBlobStatus) ID() uint32 {
	return IDClientCacheBlobStatus
}

func (pk *ClientCacheBlobStatus) Marshal(io protocol.IO) {
	count := uint32(len(pk.MissHashes))
	io.Varuint32(&count)
	protocol.FuncSliceOfLenWithMaxLen(io, count, &pk.MissHashes, io.Uint64, maxClientCacheBlobHashes)
	count = uint32(len(pk.HitHashes))
	io.Varuint32(&count)
	protocol.FuncSliceOfLenWithMaxLen(io, count, &pk.HitHashes, io.Uint64, maxClientCacheBlobHashes)
}
