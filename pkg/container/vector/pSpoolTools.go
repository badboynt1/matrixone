// Copyright 2024 Matrix Origin
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

package vector

// SetVecData is dangerous and should be used with caution.
func SetVecData(v *Vector, data []byte) {
	data = data[:cap(data)]
	v.data = data
}

// SetVecArea is dangerous and should be used with caution.
func SetVecArea(v *Vector, area []byte) {
	v.area = area
}

// GetAndClearVecData is a dangerous function that may cause data leakage.
func GetAndClearVecData(v *Vector) []byte {
	s := v.data
	v.data = nil
	return s
}

// GetAndClearVecArea is a dangerous function that may cause data leakage.
func GetAndClearVecArea(v *Vector) []byte {
	s := v.area
	v.area = nil
	return s
}
