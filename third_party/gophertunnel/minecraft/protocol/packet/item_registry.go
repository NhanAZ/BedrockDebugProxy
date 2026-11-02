package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// ItemRegistry is sent by the server to send the client a list of available items and attach client-side
// components to a custom item. This packet was formerly known as the ItemComponent packet before 1.21.60,
// which did not include item definitions but only the components.
type ItemRegistry struct {
	// Items is a list of all items with their legacy IDs which are available in the game. Failing to send any
	// of the items that are in the game will crash mobile clients. Any custom components are also attached to
	// the items in this list.
	Items []protocol.ItemEntry
}

// Current Bedrock item registries can exceed the generic protocol slice bound
// when a server exposes its full item table. Keep this field-specific bound
// finite while allowing the observed Axolotl-PM registry.
const maxItemRegistryEntries = 16 * 1024

// ID ...
func (*ItemRegistry) ID() uint32 {
	return IDItemRegistry
}

func (pk *ItemRegistry) Marshal(io protocol.IO) {
	count := uint32(len(pk.Items))
	io.Varuint32(&count)
	protocol.FuncSliceOfLenWithMaxLen(io, count, &pk.Items, func(item *protocol.ItemEntry) {
		item.Marshal(io)
	}, maxItemRegistryEntries)
}
