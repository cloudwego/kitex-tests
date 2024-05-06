# Copyright 2023 CloudWeGo Authors
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#     http://www.apache.org/licenses/LICENSE-2.0
# 
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/bin/bash

kitex -module test -gen-path old_gen test_fieldmask.thrift 
kitex -module test -thrift with_field_mask -thrift with_reflection test_fieldmask.thrift
cat test_fieldmask.thrift > test_fieldmask2.thrift
cat test_fieldmask.thrift > test_fieldmask3.thrift
kitex -module test -thrift with_field_mask -thrift field_mask_zero_required -thrift with_reflection -gen-path zero_gen test_fieldmask2.thrift
kitex -module test -thrift with_field_mask -thrift field_mask_halfway -thrift with_reflection -gen-path halfway_gen test_fieldmask3.thrift
kitex -module test -thrift with_field_mask -thrift with_reflection baseline.thrift
go mod tidy
go test ./...
