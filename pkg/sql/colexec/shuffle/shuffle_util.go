// Copyright 2021 Matrix Origin
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

package shuffle

import (
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/defines"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	plan2 "github.com/matrixorigin/matrixone/pkg/sql/plan"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

func ShuffleBlocksByHash(relData engine.RelData, newRelData engine.RelData, info engine.ShuffleBlockInfo) error {
	err := engine.ForRangeBlockInfo(1, relData.DataCnt(), relData,
		func(blk *objectio.BlockInfo) (bool, error) {
			location := blk.MetaLocation()
			objTimeStamp := location.Name()[:7]
			index := plan2.SimpleCharHashToRange(objTimeStamp, uint64(info.CNCNT))
			if info.CNIDX == int32(index) {
				newRelData.AppendBlockInfo(blk)
			}
			return true, nil
		})
	if err != nil {
		return err
	}
	return nil
}

func ShuffleBlocksByRange(proc *process.Process, relData engine.RelData, newRelData engine.RelData, info engine.ShuffleBlockInfo) error {
	var objDataMeta objectio.ObjectDataMeta
	var objMeta objectio.ObjectMeta
	var shuffleRangeUint64 []uint64
	var shuffleRangeInt64 []int64
	var init bool
	var index uint64

	err := engine.ForRangeBlockInfo(1, relData.DataCnt(), relData,
		func(blk *objectio.BlockInfo) (bool, error) {
			location := blk.MetaLocation()
			fs, err := fileservice.Get[fileservice.FileService](proc.Base.FileService, defines.SharedFileServiceName)
			if err != nil {
				return false, err
			}
			if !objectio.IsSameObjectLocVsMeta(location, objDataMeta) {
				if objMeta, err = objectio.FastLoadObjectMeta(proc.Ctx, &location, false, fs); err != nil {
					return false, err
				}
				objDataMeta = objMeta.MustDataMeta()
			}
			blkMeta := objDataMeta.GetBlockMeta(uint32(location.ID()))
			zm := blkMeta.MustGetColumn(uint16(info.ShuffleColIdx)).ZoneMap()
			if !zm.IsInited() {
				// a block with all null will send to first CN
				if info.CNIDX == 0 {
					newRelData.AppendBlockInfo(blk)
				}
			}
			if !init {
				init = true
				switch zm.GetType() {
				case types.T_int64, types.T_int32, types.T_int16:
					shuffleRangeInt64 = plan2.ShuffleRangeReEvalSigned(info.Ranges, int(info.CNCNT), info.NullCnt, info.TableCnt)
				case types.T_uint64, types.T_uint32, types.T_uint16, types.T_varchar, types.T_char, types.T_text, types.T_bit, types.T_datalink:
					shuffleRangeUint64 = plan2.ShuffleRangeReEvalUnsigned(info.Ranges, int(info.CNCNT), info.NullCnt, info.TableCnt)
				}
			}
			if shuffleRangeUint64 != nil {
				index = plan2.GetRangeShuffleIndexForZMUnsignedSlice(shuffleRangeUint64, zm)
			} else if shuffleRangeInt64 != nil {
				index = plan2.GetRangeShuffleIndexForZMSignedSlice(shuffleRangeInt64, zm)
			} else {
				index = plan2.GetRangeShuffleIndexForZM(info.ColMin, info.ColMax, zm, uint64(info.CNCNT))
			}
			if info.CNIDX == int32(index) {
				newRelData.AppendBlockInfo(blk)
			}
			return true, nil
		})

	return err

}
