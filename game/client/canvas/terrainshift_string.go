// generated by stringer -type=TerrainShift; DO NOT EDIT

package canvas

import "fmt"

const _TerrainShift_name = "TS_NORTHTS_NORTHEASTTS_EASTTS_SOUTHEASTTS_SOUTHTS_SOUTHWESTTS_WESTTS_NORTHWEST"

var _TerrainShift_index = [...]uint8{0, 8, 20, 27, 39, 47, 59, 66, 78}

func (i TerrainShift) String() string {
	if i < 0 || i+1 >= TerrainShift(len(_TerrainShift_index)) {
		return fmt.Sprintf("TerrainShift(%d)", i)
	}
	return _TerrainShift_name[_TerrainShift_index[i]:_TerrainShift_index[i+1]]
}
