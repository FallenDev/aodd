// generated by stringer -type=MoveRequestType; DO NOT EDIT

package game

import "fmt"

const _MoveRequestType_name = "MR_ERRORMR_MOVEMR_MOVE_CANCELMR_SIZE"

var _MoveRequestType_index = [...]uint8{0, 8, 15, 29, 36}

func (i MoveRequestType) String() string {
	if i < 0 || i+1 >= MoveRequestType(len(_MoveRequestType_index)) {
		return fmt.Sprintf("MoveRequestType(%d)", i)
	}
	return _MoveRequestType_name[_MoveRequestType_index[i]:_MoveRequestType_index[i+1]]
}
