// generated by stringer -type=Direction; DO NOT EDIT

package coord

import "fmt"

const _Direction_name = "NorthEastSouthWest"

var _Direction_index = [...]uint8{0, 5, 9, 14, 18}

func (i Direction) String() string {
	if i+1 >= Direction(len(_Direction_index)) {
		return fmt.Sprintf("Direction(%d)", i)
	}
	return _Direction_name[_Direction_index[i]:_Direction_index[i+1]]
}
