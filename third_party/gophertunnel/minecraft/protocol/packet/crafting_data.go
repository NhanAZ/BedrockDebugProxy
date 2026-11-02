package packet

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// CraftingData is sent by the server to let the client know all crafting data that the server maintains. This
// includes shapeless crafting, crafting table recipes, furnace recipes etc. Each crafting station's recipes
// are included in it.
type CraftingData struct {
	// ShapedRecipes through SmithingTrimRecipes are the typed recipe vectors used by 1.26.40.
	ShapedRecipes             []protocol.ShapedRecipe
	ShapelessRecipes          []protocol.ShapelessRecipe
	MultiRecipes              []protocol.MultiRecipe
	UserDataShapelessRecipes  []protocol.UserDataShapelessRecipe
	ShapelessChemistryRecipes []protocol.ShapelessChemistryRecipe
	ShapedChemistryRecipes    []protocol.ShapedChemistryRecipe
	SmithingTransformRecipes  []protocol.SmithingTransformRecipe
	SmithingTrimRecipes       []protocol.SmithingTrimRecipe
	// PotionRecipes is a list of all potion mixing recipes which may be used in the brewing stand.
	PotionRecipes []protocol.PotionRecipe
	// PotionContainerChangeRecipes is a list of all recipes to convert a potion from one type to another,
	// such as from a drinkable potion to a splash potion, or from a splash potion to a lingering potion.
	PotionContainerChangeRecipes []protocol.PotionContainerChangeRecipe
	// MaterialReducers is a list of all material reducers which is used in education edition chemistry.
	MaterialReducers []protocol.MaterialReducer
	// ClearRecipes indicates if all recipes currently active on the client should be cleaned. Doing this
	// means that the client will have no recipes active by itself: Any CraftingData packets previously sent
	// will also be discarded, and only the recipes in this CraftingData packet will be used.
	ClearRecipes bool
}

// Current Bedrock servers can expose more than the generic protocol slice
// bound in one crafting table. Keep each recipe vector bounded while allowing
// the observed Axolotl-PM recipe set.
const maxCraftingDataEntries = 16 * 1024

// ID ...
func (*CraftingData) ID() uint32 {
	return IDCraftingData
}

func (pk *CraftingData) Marshal(io protocol.IO) {
	marshalCraftingSlice(io, &pk.ShapedRecipes, func(recipe *protocol.ShapedRecipe) { recipe.Marshal(io) })
	marshalCraftingSlice(io, &pk.ShapelessRecipes, func(recipe *protocol.ShapelessRecipe) { recipe.Marshal(io) })
	marshalCraftingSlice(io, &pk.MultiRecipes, func(recipe *protocol.MultiRecipe) { recipe.Marshal(io) })
	marshalCraftingSlice(io, &pk.UserDataShapelessRecipes, func(recipe *protocol.UserDataShapelessRecipe) { recipe.Marshal(io) })
	marshalCraftingSlice(io, &pk.ShapelessChemistryRecipes, func(recipe *protocol.ShapelessChemistryRecipe) { recipe.Marshal(io) })
	marshalCraftingSlice(io, &pk.ShapedChemistryRecipes, func(recipe *protocol.ShapedChemistryRecipe) { recipe.Marshal(io) })
	marshalCraftingSlice(io, &pk.SmithingTransformRecipes, func(recipe *protocol.SmithingTransformRecipe) { recipe.Marshal(io) })
	marshalCraftingSlice(io, &pk.SmithingTrimRecipes, func(recipe *protocol.SmithingTrimRecipe) { recipe.Marshal(io) })
	marshalCraftingSlice(io, &pk.PotionRecipes, func(recipe *protocol.PotionRecipe) { recipe.Marshal(io) })
	marshalCraftingSlice(io, &pk.PotionContainerChangeRecipes, func(recipe *protocol.PotionContainerChangeRecipe) { recipe.Marshal(io) })
	marshalCraftingSlice(io, &pk.MaterialReducers, io.MaterialReducer)
	io.Bool(&pk.ClearRecipes)
}

func marshalCraftingSlice[T any, S ~*[]T](io protocol.IO, values S, marshal func(*T)) {
	count := uint32(len(*values))
	io.Varuint32(&count)
	protocol.FuncSliceOfLenWithMaxLen(io, count, values, marshal, maxCraftingDataEntries)
}
