// Code generated by "stringer -type CardName"; DO NOT EDIT.

package game

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Librarian-0]
	_ = x[Magician-1]
	_ = x[Shieldmancer-2]
	_ = x[MindMage-3]
	_ = x[Angel-4]
	_ = x[Pyromancer-5]
	_ = x[Bloodeater-6]
	_ = x[Conjurer-7]
	_ = x[Mortician-8]
}

const _CardName_name = "LibrarianMagicianShieldmancerMindMageAngelPyromancerBloodeaterConjurerMortician"

var _CardName_index = [...]uint8{0, 9, 17, 29, 37, 42, 52, 62, 70, 79}

func (i CardName) String() string {
	if i < 0 || i >= CardName(len(_CardName_index)-1) {
		return "CardName(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _CardName_name[_CardName_index[i]:_CardName_index[i+1]]
}
