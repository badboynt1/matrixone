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

package vm

// var stringFunc = [...]func(any, *bytes.Buffer){
// 	Top:         top.String,
// 	Join:        join.String,
// 	Semi:        semi.String,
// 	RightSemi:   rightsemi.String,
// 	RightAnti:   rightanti.String,
// 	Left:        left.String,
// 	Right:       right.String,
// 	Single:      single.String,
// 	Limit:       limit.String,
// 	Order:       order.String,
// 	Group:       group.String,
// 	Window:      window.String,
// 	Merge:       merge.String,
// 	Output:      output.String,
// 	Offset:      offset.String,
// 	Product:     product.String,
// 	Restrict:    restrict.String,
// 	Dispatch:    dispatch.String,
// 	Connector:   connector.String,
// 	Projection:  projection.String,
// 	Anti:        anti.String,
// 	Mark:        mark.String,
// 	MergeBlock:  mergeblock.String,
// 	MergeDelete: mergedelete.String,
// 	LoopJoin:    loopjoin.String,
// 	LoopLeft:    loopleft.String,
// 	LoopSingle:  loopsingle.String,
// 	LoopSemi:    loopsemi.String,
// 	LoopAnti:    loopanti.String,
// 	LoopMark:    loopmark.String,

// 	MergeTop:       mergetop.String,
// 	MergeLimit:     mergelimit.String,
// 	MergeOrder:     mergeorder.String,
// 	MergeGroup:     mergegroup.String,
// 	MergeOffset:    mergeoffset.String,
// 	MergeRecursive: mergerecursive.String,
// 	MergeCTE:       mergecte.String,

// 	Deletion:        deletion.String,
// 	Insert:          insert.String,
//  FuzzyFilter:     fuzzyfilter.String,
// 	OnDuplicateKey:  onduplicatekey.String,
// 	PreInsert:       preinsert.String,
// 	PreInsertUnique: preinsertunique.String,
// 	External:        external.String,

// 	Minus:        minus.String,
// 	Intersect:    intersect.String,
// 	IntersectAll: intersectall.String,

// 	HashBuild: hashbuild.String,

// 	TableFunction: table_function.String,

// 	LockOp: lockop.String,

// 	Shuffle: shuffle.String,
// 	Stream:  stream.String,
// }

// var prepareFunc = [...]func(*process.Process, any) error{
// 	Top:         top.Prepare,
// 	Join:        join.Prepare,
// 	Semi:        semi.Prepare,
// 	RightSemi:   rightsemi.Prepare,
// 	RightAnti:   rightanti.Prepare,
// 	Left:        left.Prepare,
// 	Right:       right.Prepare,
// 	Single:      single.Prepare,
// 	Limit:       limit.Prepare,
// 	Order:       order.Prepare,
// 	Group:       group.Prepare,
// 	Window:      window.Prepare,
// 	Merge:       merge.Prepare,
// 	Output:      output.Prepare,
// 	Offset:      offset.Prepare,
// 	Product:     product.Prepare,
// 	Restrict:    restrict.Prepare,
// 	Dispatch:    dispatch.Prepare,
// 	Connector:   connector.Prepare,
// 	Projection:  projection.Prepare,
// 	Anti:        anti.Prepare,
// 	Mark:        mark.Prepare,
// 	MergeBlock:  mergeblock.Prepare,
// 	MergeDelete: mergedelete.Prepare,
// 	LoopJoin:    loopjoin.Prepare,
// 	LoopLeft:    loopleft.Prepare,
// 	LoopSingle:  loopsingle.Prepare,
// 	LoopSemi:    loopsemi.Prepare,
// 	LoopAnti:    loopanti.Prepare,
// 	LoopMark:    loopmark.Prepare,

// 	MergeTop:       mergetop.Prepare,
// 	MergeLimit:     mergelimit.Prepare,
// 	MergeOrder:     mergeorder.Prepare,
// 	MergeGroup:     mergegroup.Prepare,
// 	MergeOffset:    mergeoffset.Prepare,
// 	MergeRecursive: mergerecursive.Prepare,
// 	MergeCTE:       mergecte.Prepare,

// 	Deletion:        deletion.Prepare,
// 	Insert:          insert.Prepare,
//  FuzzyFilter:     fuzzyfilter.Prepare,
// 	OnDuplicateKey:  onduplicatekey.Prepare,
// 	PreInsert:       preinsert.Prepare,
// 	PreInsertUnique: preinsertunique.Prepare,
// 	External:        external.Prepare,

// 	Minus:        minus.Prepare,
// 	Intersect:    intersect.Prepare,
// 	IntersectAll: intersectall.Prepare,

// 	HashBuild: hashbuild.Prepare,

// 	TableFunction: table_function.Prepare,

// 	LockOp: lockop.Prepare,

// 	Shuffle: shuffle.Prepare,
// 	Stream:  stream.Prepare,
// }

// var execFunc = [...]func(int, *process.Process, any, bool, bool) (process.ExecStatus, error){
// 	Top:         top.Call,
// 	Join:        join.Call,
// 	Semi:        semi.Call,
// 	RightSemi:   rightsemi.Call,
// 	RightAnti:   rightanti.Call,
// 	Left:        left.Call,
// 	Right:       right.Call,
// 	Single:      single.Call,
// 	Limit:       limit.Call,
// 	Order:       order.Call,
// 	Group:       group.Call,
// 	Window:      window.Call,
// 	Merge:       merge.Call,
// 	Output:      output.Call,
// 	Offset:      offset.Call,
// 	Product:     product.Call,
// 	Restrict:    restrict.Call,
// 	Dispatch:    dispatch.Call,
// 	Connector:   connector.Call,
// 	Projection:  projection.Call,
// 	Anti:        anti.Call,
// 	Mark:        mark.Call,
// 	MergeBlock:  mergeblock.Call,
// 	MergeDelete: mergedelete.Call,
// 	LoopJoin:    loopjoin.Call,
// 	LoopLeft:    loopleft.Call,
// 	LoopSingle:  loopsingle.Call,
// 	LoopSemi:    loopsemi.Call,
// 	LoopAnti:    loopanti.Call,
// 	LoopMark:    loopmark.Call,

// 	MergeTop:       mergetop.Call,
// 	MergeLimit:     mergelimit.Call,
// 	MergeOrder:     mergeorder.Call,
// 	MergeGroup:     mergegroup.Call,
// 	MergeOffset:    mergeoffset.Call,
// 	MergeRecursive: mergerecursive.Call,
// 	MergeCTE:       mergecte.Call,

// 	Deletion: deletion.Call,
// 	Insert:   insert.Call,
//  FuzzyFilter: fuzzyfilter.Call,
// 	External: external.Call,

// 	OnDuplicateKey:  onduplicatekey.Call,
// 	PreInsert:       preinsert.Call,
// 	PreInsertUnique: preinsertunique.Call,

// 	Minus:        minus.Call,
// 	Intersect:    intersect.Call,
// 	IntersectAll: intersectall.Call,

// 	HashBuild: hashbuild.Call,

// 	TableFunction: table_function.Call,

// 	LockOp: lockop.Call,

// 	Shuffle: shuffle.Call,
// 	Stream:  stream.Call,
// }
