// Copyright 2022 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plan

import (
	"github.com/matrixorigin/matrixone/pkg/container/hashtable"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
)

const (
	HashMapSizeForShuffle             = 250000
	MAXShuffleDOP                     = 64
	ShuffleThreshHold                 = 50000
	ShuffleTypeThreshHold             = 32
	ShuffleTypeThreshHoldForRightJoin = 512
)

const (
	ShuffleToRegIndex        int32 = 0
	ShuffleToLocalMatchedReg int32 = 1
	ShuffleToMultiMatchedReg int32 = 2
)

func SimpleCharHashToRange(bytes []byte, upperLimit uint64) uint64 {
	lenBytes := len(bytes)
	//sample five bytes
	h := (uint64(bytes[0])*(uint64(bytes[lenBytes/4])+uint64(bytes[lenBytes/2])+uint64(bytes[lenBytes*3/4])) + uint64(bytes[lenBytes-1]))
	return hashtable.Int64HashWithFixedSeed(h) % upperLimit
}

func SimpleInt64HashToRange(i uint64, upperLimit uint64) uint64 {
	return hashtable.Int64HashWithFixedSeed(i) % upperLimit
}

func GetRangeShuffleIndexForZM(minVal, maxVal int64, zm objectio.ZoneMap, upplerLimit uint64) uint64 {
	switch zm.GetType() {
	case types.T_int64:
		return GetRangeShuffleIndexSigned(minVal, maxVal, types.DecodeInt64(zm.GetMinBuf()), upplerLimit)
	case types.T_int32:
		return GetRangeShuffleIndexSigned(minVal, maxVal, int64(types.DecodeInt32(zm.GetMinBuf())), upplerLimit)
	case types.T_int16:
		return GetRangeShuffleIndexSigned(minVal, maxVal, int64(types.DecodeInt16(zm.GetMinBuf())), upplerLimit)
	case types.T_uint64:
		return GetRangeShuffleIndexUnsigned(uint64(minVal), uint64(maxVal), types.DecodeUint64(zm.GetMinBuf()), upplerLimit)
	case types.T_uint32:
		return GetRangeShuffleIndexUnsigned(uint64(minVal), uint64(maxVal), uint64(types.DecodeUint32(zm.GetMinBuf())), upplerLimit)
	case types.T_uint16:
		return GetRangeShuffleIndexUnsigned(uint64(minVal), uint64(maxVal), uint64(types.DecodeUint16(zm.GetMinBuf())), upplerLimit)
	}
	panic("unsupported shuffle type!")
}

func GetRangeShuffleIndexSigned(minVal, maxVal, currentVal int64, upplerLimit uint64) uint64 {
	if currentVal <= minVal {
		return 0
	} else if currentVal >= maxVal {
		return upplerLimit - 1
	} else {
		step := uint64(maxVal-minVal) / upplerLimit
		ret := uint64(currentVal-minVal) / step
		if ret >= upplerLimit {
			return upplerLimit - 1
		}
		return ret
	}
}

func GetRangeShuffleIndexUnsigned(minVal, maxVal, currentVal uint64, upplerLimit uint64) uint64 {
	if currentVal <= minVal {
		return 0
	} else if currentVal >= maxVal {
		return upplerLimit - 1
	} else {
		step := (maxVal - minVal) / upplerLimit
		ret := (currentVal - minVal) / step
		if ret >= upplerLimit {
			return upplerLimit - 1
		}
		return ret
	}
}

func GetHashColumn(expr *plan.Expr) (*plan.ColRef, int32) {
	switch exprImpl := expr.Expr.(type) {
	case *plan.Expr_F:
		for _, arg := range exprImpl.F.Args {
			col, typ := GetHashColumn(arg)
			if col != nil {
				return col, typ
			}
		}
	case *plan.Expr_Col:
		return exprImpl.Col, expr.Typ.Id
	}
	return nil, -1
}

func maybeSorted(n *plan.Node, builder *QueryBuilder, tag int32) bool {
	// for scan node, primary key and cluster by may be sorted
	if n.NodeType == plan.Node_TABLE_SCAN {
		return n.BindingTags[0] == tag
	}
	// for inner join, if left child may be sorted, then inner join may be sorted
	if n.NodeType == plan.Node_JOIN && n.JoinType == plan.Node_INNER {
		leftChild := builder.qry.Nodes[n.Children[0]]
		return maybeSorted(leftChild, builder, tag)
	}
	return false
}

func determinShuffleType(col *plan.ColRef, n *plan.Node, builder *QueryBuilder) {
	// hash by default
	n.Stats.HashmapStats.ShuffleType = plan.ShuffleType_Hash

	if builder == nil {
		return
	}
	tableDef, ok := builder.tag2Table[col.RelPos]
	if !ok {
		return
	}
	colName := tableDef.Cols[col.ColPos].Name

	// for shuffle join, if left child is not sorted, the cost will be very high
	// should use complex shuffle type
	if n.NodeType == plan.Node_JOIN {
		leftSorted := true
		if GetSortOrder(tableDef, colName) != 0 {
			leftSorted = false
		}
		if !maybeSorted(builder.qry.Nodes[n.Children[0]], builder, col.RelPos) {
			leftSorted = false
		}
		if !leftSorted {
			if n.BuildOnLeft {
				if builder.qry.Nodes[n.Children[0]].Stats.Outcnt > ShuffleTypeThreshHoldForRightJoin*builder.qry.Nodes[n.Children[1]].Stats.Outcnt {
					n.Stats.HashmapStats.ShuffleTypeForMultiCN = plan.ShuffleTypeForMultiCN_Complex
				}
			}
			if builder.qry.Nodes[n.Children[0]].Stats.Outcnt > ShuffleTypeThreshHold*builder.qry.Nodes[n.Children[1]].Stats.Outcnt {
				n.Stats.HashmapStats.ShuffleTypeForMultiCN = plan.ShuffleTypeForMultiCN_Complex
			}
		}
	}

	sc := builder.compCtx.GetStatsCache()
	if sc == nil {
		return
	}
	s := sc.GetStatsInfoMap(tableDef.TblId)
	n.Stats.HashmapStats.ShuffleType = plan.ShuffleType_Range
	n.Stats.HashmapStats.ShuffleColMin = int64(s.MinValMap[colName])
	n.Stats.HashmapStats.ShuffleColMax = int64(s.MaxValMap[colName])

}

// to determine if join need to go shuffle
func determinShuffleForJoin(n *plan.Node, builder *QueryBuilder) {
	// do not shuffle by default
	n.Stats.HashmapStats.ShuffleColIdx = -1
	if n.NodeType != plan.Node_JOIN {
		return
	}
	switch n.JoinType {
	case plan.Node_INNER, plan.Node_ANTI, plan.Node_SEMI, plan.Node_LEFT, plan.Node_RIGHT:
	default:
		return
	}

	// for now, if join children is agg, do not allow shuffle
	if builder.qry.Nodes[n.Children[0]].NodeType == plan.Node_AGG || builder.qry.Nodes[n.Children[1]].NodeType == plan.Node_AGG {
		return
	}

	if n.Stats.HashmapStats.HashmapSize < HashMapSizeForShuffle {
		return
	}
	idx := 0
	if !builder.IsEquiJoin(n) {
		return
	}
	leftTags := make(map[int32]any)
	for _, tag := range builder.enumerateTags(n.Children[0]) {
		leftTags[tag] = nil
	}
	rightTags := make(map[int32]any)
	for _, tag := range builder.enumerateTags(n.Children[1]) {
		rightTags[tag] = nil
	}
	// for now ,only support the first join condition
	for i := range n.OnList {
		if isEquiCond(n.OnList[i], leftTags, rightTags) {
			idx = i
			break
		}
	}

	//find the highest ndv
	highestNDV := n.OnList[idx].Ndv
	if highestNDV < ShuffleThreshHold {
		return
	}

	// get the column of left child
	hashCol, typ := GetHashColumn(n.OnList[idx])
	if hashCol == nil {
		return
	}
	//for now ,only support integer and string type
	switch types.T(typ) {
	case types.T_int64, types.T_int32, types.T_int16, types.T_uint64, types.T_uint32, types.T_uint16:
		n.Stats.HashmapStats.ShuffleColIdx = int32(idx)
		n.Stats.HashmapStats.Shuffle = true
		determinShuffleType(hashCol, n, builder)
	case types.T_varchar, types.T_char, types.T_text:
		n.Stats.HashmapStats.ShuffleColIdx = int32(idx)
		n.Stats.HashmapStats.Shuffle = true
	}

	// for now, do not support hash shuffle join. will support it in the future
	if n.Stats.HashmapStats.ShuffleType == plan.ShuffleType_Hash {
		n.Stats.HashmapStats.Shuffle = false
	}

}

// to determine if groupby need to go shuffle
func determinShuffleForGroupBy(n *plan.Node, builder *QueryBuilder) {
	// do not shuffle by default
	n.Stats.HashmapStats.ShuffleColIdx = -1

	if n.NodeType != plan.Node_AGG {
		return
	}
	if len(n.GroupBy) == 0 {
		return
	}

	// for now, if agg children is agg, do not allow shuffle
	if builder.qry.Nodes[n.Children[0]].NodeType == plan.Node_AGG {
		return
	}

	if n.Stats.HashmapStats.HashmapSize < HashMapSizeForShuffle {
		return
	}
	//find the highest ndv
	highestNDV := n.GroupBy[0].Ndv
	idx := 0
	for i := range n.GroupBy {
		if n.GroupBy[i].Ndv > highestNDV {
			highestNDV = n.GroupBy[i].Ndv
			idx = i
		}
	}
	if highestNDV < ShuffleThreshHold {
		return
	}

	hashCol, typ := GetHashColumn(n.GroupBy[idx])
	if hashCol == nil {
		return
	}
	//for now ,only support integer and string type
	switch types.T(typ) {
	case types.T_int64, types.T_int32, types.T_int16, types.T_uint64, types.T_uint32, types.T_uint16:
		n.Stats.HashmapStats.ShuffleColIdx = int32(idx)
		n.Stats.HashmapStats.Shuffle = true
		determinShuffleType(hashCol, n, builder)
	case types.T_varchar, types.T_char, types.T_text:
		n.Stats.HashmapStats.ShuffleColIdx = int32(idx)
		n.Stats.HashmapStats.Shuffle = true
	}

	//shuffle join-> shuffle group ,if they use the same hask key, the group can reuse the shuffle method
	child := builder.qry.Nodes[n.Children[0]]
	if child.NodeType == plan.Node_JOIN {
		if n.Stats.HashmapStats.Shuffle && child.Stats.HashmapStats.Shuffle {
			// shuffle group can follow shuffle join
			if n.Stats.HashmapStats.ShuffleType == child.Stats.HashmapStats.ShuffleType {
				groupHashCol, _ := GetHashColumn(n.GroupBy[n.Stats.HashmapStats.ShuffleColIdx])
				switch exprImpl := child.OnList[child.Stats.HashmapStats.ShuffleColIdx].Expr.(type) {
				case *plan.Expr_F:
					for _, arg := range exprImpl.F.Args {
						joinHashCol, _ := GetHashColumn(arg)
						if groupHashCol.RelPos == joinHashCol.RelPos && groupHashCol.ColPos == joinHashCol.ColPos {
							n.Stats.HashmapStats.ShuffleMethod = plan.ShuffleMethod_Reuse
							return
						}
					}
				}
			}
			// shuffle group can not follow shuffle join, need to reshuffle
			n.Stats.HashmapStats.ShuffleMethod = plan.ShuffleMethod_Reshuffle
		}
	}

}

func GetShuffleDop() (dop int) {
	return MAXShuffleDOP
}

// default shuffle type for scan is hash
// for table with primary key, and ndv of first column in primary key is high enough, use range shuffle
// only support integer type
func determinShuffleForScan(n *plan.Node, builder *QueryBuilder) {
	if n.Stats.HashmapStats == nil {
		n.Stats.HashmapStats = &plan.HashMapStats{}
	}
	n.Stats.HashmapStats.Shuffle = true
	n.Stats.HashmapStats.ShuffleType = plan.ShuffleType_Hash
	if n.TableDef.Pkey != nil {
		firstColName := n.TableDef.Pkey.Names[0]
		firstColID := n.TableDef.Name2ColIndex[firstColName]
		sc := builder.compCtx.GetStatsCache()
		if sc == nil {
			return
		}
		s := sc.GetStatsInfoMap(n.TableDef.TblId)
		if s.NdvMap[firstColName] < ShuffleThreshHold {
			return
		}
		switch types.T(n.TableDef.Cols[firstColID].Typ.Id) {
		case types.T_int64, types.T_int32, types.T_int16, types.T_uint64, types.T_uint32, types.T_uint16:
			n.Stats.HashmapStats.ShuffleType = plan.ShuffleType_Range
			n.Stats.HashmapStats.ShuffleColIdx = int32(n.TableDef.Cols[firstColID].Seqnum)
			n.Stats.HashmapStats.ShuffleColMin = int64(s.MinValMap[firstColName])
			n.Stats.HashmapStats.ShuffleColMax = int64(s.MaxValMap[firstColName])
		}
	}
}

func determineShuffleMethod(nodeID int32, builder *QueryBuilder) {
	node := builder.qry.Nodes[nodeID]
	if len(node.Children) > 0 {
		for _, child := range node.Children {
			determineShuffleMethod(child, builder)
		}
	}
	switch node.NodeType {
	case plan.Node_AGG:
		determinShuffleForGroupBy(node, builder)
	case plan.Node_TABLE_SCAN:
		determinShuffleForScan(node, builder)
	case plan.Node_JOIN:
		determinShuffleForJoin(node, builder)
	default:
	}

}
