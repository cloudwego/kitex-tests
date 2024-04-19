# Copyright 2023 CloudWeGo Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

  if command -v thriftgo >/dev/null 2>&1; then
      echo thriftgo:
      thriftgo -version
  else
      echo "安装 thriftgo 失败，请检查错误信息并重试"
      exit 1
  fi

  if command -v kitex >/dev/null 2>&1; then
      echo kitex:
      kitex -version
  else
      echo "安装 kitex 失败，请检查错误信息并重试"
      exit 1
  fi