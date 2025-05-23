package export

import "github.com/elastic/go-elasticsearch/v8/typedapi/types"

// This is provide all the types that are part of the union.
type _sortCombinations struct {
	v types.SortCombinations
}

func NewSortCombinations() *_sortCombinations {
	return &_sortCombinations{v: nil}
}

func (u *_sortCombinations) Field(field string) *_sortCombinations {

	u.v = &field

	return u
}

func (u *_sortCombinations) SortOptions(sortoptions types.SortOptionsVariant) *_sortCombinations {

	u.v = &sortoptions

	return u
}

// Interface implementation for SortOptions in SortCombinations union
func (u *_sortOptions) SortCombinationsCaster() *types.SortCombinations {
	t := types.SortCombinations(u.v)
	return &t
}

func (u *_sortCombinations) SortCombinationsCaster() *types.SortCombinations {
	return &u.v
}
