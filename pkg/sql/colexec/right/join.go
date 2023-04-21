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

package right

import (
	"bytes"

	"github.com/matrixorigin/matrixone/pkg/common/hashmap"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec"
	"github.com/matrixorigin/matrixone/pkg/sql/plan"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

func String(_ any, buf *bytes.Buffer) {
	buf.WriteString(" right join ")
}

func Prepare(proc *process.Process, arg any) error {
	ap := arg.(*Argument)
	ap.ctr = new(container)
	ap.ctr.inBuckets = make([]uint8, hashmap.UnitLimit)
	ap.ctr.evecs = make([]evalVector, len(ap.Conditions[0]))
	ap.ctr.vecs = make([]*vector.Vector, len(ap.Conditions[0]))
	ap.ctr.bat = batch.NewWithSize(len(ap.RightTypes))
	ap.ctr.bat.Zs = proc.Mp().GetSels()
	for i, typ := range ap.RightTypes {
		ap.ctr.bat.Vecs[i] = vector.NewVec(typ)
	}
	return nil
}

func Call(idx int, proc *process.Process, arg any, isFirst bool, isLast bool) (bool, error) {
	analyze := proc.GetAnalyze(idx)
	analyze.Start()
	defer analyze.Stop()
	ap := arg.(*Argument)
	ctr := ap.ctr
	for {
		switch ctr.state {
		case Build:
			if err := ctr.build(ap, proc, analyze); err != nil {
				ap.Free(proc, true)
				return false, err
			}
			ctr.state = Probe

		case Probe:
			bat := <-proc.Reg.MergeReceivers[0].Ch
			if bat == nil {
				ctr.state = SendLast
				continue
			}
			if bat.Length() == 0 {
				continue
			}

			if ctr.bat == nil || ctr.bat.Length() == 0 {
				bat.Clean(proc.Mp())
				continue
			}

			if err := ctr.probe(bat, ap, proc, analyze, isFirst, isLast); err != nil {
				ap.Free(proc, true)
				return false, err
			}

			return false, nil
		case SendLast:
			if ctr.bat == nil || ctr.bat.Length() == 0 {
				ctr.state = End
			} else {
				setNil, err := ctr.sendLast(ap, proc, analyze, isFirst, isLast)

				if err != nil {
					ap.Free(proc, true)
					return false, err
				}
				ctr.state = End
				if setNil {
					continue
				}

				return false, nil
			}

		default:
			proc.SetInputBatch(nil)
			ap.Free(proc, false)
			return true, nil
		}
	}
}

func (ctr *container) build(ap *Argument, proc *process.Process, analyze process.Analyze) error {
	bat := <-proc.Reg.MergeReceivers[1].Ch
	if bat != nil {
		ctr.bat = bat
		ctr.mp = bat.Ht.(*hashmap.JoinMap).Dup()
		ctr.matched = make([]uint8, bat.Length())
		analyze.Alloc(ctr.mp.Map().Size())
	}
	return nil
}

func (ctr *container) sendLast(ap *Argument, proc *process.Process, analyze process.Analyze, isFirst bool, isLast bool) (bool, error) {
	if !ap.IsMerger {
		ap.Channel <- &ctr.matched
		//goto ctr.end directly
		return true, nil
	}

	if ap.NumCPU > 1 {
		cnt := 1
		for v := range ap.Channel {
			for i, val := range *v {
				ctr.matched[i] |= val
			}
			cnt++
			if cnt == int(ap.NumCPU) {
				close(ap.Channel)
				break
			}
		}
	}

	count := ctr.bat.Length()
	unmatched := 0
	for i := 0; i < count; i++ {
		ctr.matched[i] = 1 - ctr.matched[i]
		unmatched += int(ctr.matched[i])
	}

	if unmatched == 0 {
		return true, nil
	}

	rbat := batch.NewWithSize(len(ap.Result))
	rbat.Zs = proc.Mp().GetSels()

	for i, rp := range ap.Result {
		if rp.Rel == 0 {
			rbat.Vecs[i] = proc.GetVector(ap.LeftTypes[rp.Pos])
		} else {
			rbat.Vecs[i] = proc.GetVector(ap.RightTypes[rp.Pos])
		}
	}

	for j, rp := range ap.Result {
		if rp.Rel == 0 {
			if err := vector.AppendMultiFixed(rbat.Vecs[j], 0, true, unmatched, proc.Mp()); err != nil {
				rbat.Clean(proc.Mp())
				return false, err
			}
		} else {
			if err := rbat.Vecs[j].UnionBatch(ctr.bat.Vecs[rp.Pos], 0, count, ctr.matched, proc.Mp()); err != nil {
				rbat.Clean(proc.Mp())
				return false, err
			}
		}

	}
	for i := 0; i < count; i++ {
		if ctr.matched[i] == 1 {
			rbat.Zs = append(rbat.Zs, ctr.bat.Zs[i])
		}
	}
	analyze.Output(rbat, isLast)
	proc.SetInputBatch(rbat)
	return false, nil
}

func (ctr *container) probe(bat *batch.Batch, ap *Argument, proc *process.Process, anal process.Analyze, isFirst bool, isLast bool) error {
	defer proc.PutBatch(bat)
	anal.Input(bat, isFirst)
	rbat := batch.NewWithSize(len(ap.Result))
	rbat.Zs = proc.Mp().GetSels()
	for i, rp := range ap.Result {
		if rp.Rel == 0 {
			rbat.Vecs[i] = proc.GetVector(*bat.Vecs[rp.Pos].GetType())
		} else {
			rbat.Vecs[i] = proc.GetVector(*ctr.bat.Vecs[rp.Pos].GetType())
		}
	}

	if err := ctr.evalJoinCondition(bat, ap.Conditions[0], proc); err != nil {
		return err
	}
	defer ctr.cleanEvalVectors(proc.Mp())

	count := bat.Length()
	mSels := ctr.mp.Sels()
	itr := ctr.mp.Map().NewIterator()
	for i := 0; i < count; i += hashmap.UnitLimit {
		n := count - i
		if n > hashmap.UnitLimit {
			n = hashmap.UnitLimit
		}
		copy(ctr.inBuckets, hashmap.OneUInt8s)
		vals, zvals := itr.Find(i, n, ctr.vecs, ctr.inBuckets)
		for k := 0; k < n; k++ {
			if ctr.inBuckets[k] == 0 || zvals[k] == 0 || vals[k] == 0 {
				continue
			}
			sels := mSels[vals[k]-1]
			if ap.Cond != nil {
				for _, sel := range sels {
					vec, err := colexec.JoinFilterEvalExprInBucket(bat, ctr.bat, i+k, int(sel), proc, ap.Cond)
					if err != nil {
						return err
					}
					if vec.IsConstNull() || vec.GetNulls().Contains(0) {
						vec.Free(proc.Mp())
						continue
					}
					bs := vector.MustFixedCol[bool](vec)
					if !bs[0] {
						vec.Free(proc.Mp())
						continue
					}
					vec.Free(proc.Mp())
					for j, rp := range ap.Result {
						if rp.Rel == 0 {
							if err := rbat.Vecs[j].UnionOne(bat.Vecs[rp.Pos], int64(i+k), proc.Mp()); err != nil {
								rbat.Clean(proc.Mp())
								return err
							}
						} else {
							if err := rbat.Vecs[j].UnionOne(ctr.bat.Vecs[rp.Pos], int64(sel), proc.Mp()); err != nil {
								rbat.Clean(proc.Mp())
								return err
							}
						}
					}
					ctr.matched[sel] = 1
					rbat.Zs = append(rbat.Zs, ctr.bat.Zs[sel])
				}
			} else {
				for j, rp := range ap.Result {
					if rp.Rel == 0 {
						if err := rbat.Vecs[j].UnionMulti(bat.Vecs[rp.Pos], int64(i+k), len(sels), proc.Mp()); err != nil {
							rbat.Clean(proc.Mp())
							return err
						}
					} else {
						if err := rbat.Vecs[j].Union(ctr.bat.Vecs[rp.Pos], sels, proc.Mp()); err != nil {
							rbat.Clean(proc.Mp())
							return err
						}
					}
				}
				for _, sel := range sels {
					ctr.matched[sel] = 1
					rbat.Zs = append(rbat.Zs, ctr.bat.Zs[sel])
				}
			}
		}
	}
	proc.SetInputBatch(rbat)
	return nil
}

func (ctr *container) evalJoinCondition(bat *batch.Batch, conds []*plan.Expr, proc *process.Process) error {
	for i, cond := range conds {
		vec, err := colexec.EvalExpr(bat, proc, cond)
		if err != nil {
			ctr.cleanEvalVectors(proc.Mp())
			return err
		}
		ctr.vecs[i] = vec
		ctr.evecs[i].vec = vec
		ctr.evecs[i].needFree = true
		for j := range bat.Vecs {
			if bat.Vecs[j] == vec {
				ctr.evecs[i].needFree = false
				break
			}
		}
	}
	return nil
}
