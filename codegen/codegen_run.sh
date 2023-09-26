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

#!/bin/bash

function clean_codegen {
  rm -rf ./kitex_gen && rm -rf ./script && rm -rf main.go && rm -rf handler.go && rm -rf kitex_info.yaml && rm -rf build.sh && rm -rf ./conf
}

function check_cmd {

    local filename=$1

    test_cmds=(
      "kitex -module codegen-test $filename"
      "kitex -module codegen-test -service a.b.c $filename"
      "kitex -module codegen-test -combine-service $filename"
      "kitex -module codegen-test -thrift template=slim $filename"
      "kitex -module codegen-test -thrift keep_unknown_fields $filename"
      "kitex -module codegen-test -thrift template=slim -thrift keep_unknown_fields $filename"
    )

    skip_error_info=(
    "no service defined in the IDL"
    )

    for cmd in "${test_cmds[@]}"; do
      clean_codegen
      # 验证代码生成
      local tmp_file=$(mktemp) # 创建一个临时文件来存储输出
      if ! eval "$cmd" > "$tmp_file" 2>&1; then
        if ! grep -q -E "$(printf '%s\n' "${skip_error_info[@]}" | paste -sd '|' -)" "$tmp_file"; then
          echo "$cmd" >> "errors.txt"
          cat "$tmp_file" >> "errors.txt" # 将错误输出添加到错误文件中
          echo "Error: $cmd"
        fi
        rm "$tmp_file" # 删除临时文件
        continue
      fi

      rm "$tmp_file" # 删除临时文件

      # go mod 不展示输出，会干扰看结果，如果这一步出问题了，下一步 go build 会报错，所以不用担心
      go mod tidy > /dev/null 2>&1

      # 验证编译
      local tmp_file_2=$(mktemp) # 创建一个临时文件来存储输出
      if ! eval "go build ./..." > "$tmp_file_2" 2>&1; then
        echo "$cmd" >> "errors.txt"
        cat "$tmp_file_2" >> "errors.txt" # 将错误输出添加到错误文件中
        echo "Error: $cmd"
      fi
      rm "$tmp_file_2" # 删除临时文件
      done
}

function main {

  if [ -e "errors.txt" ]; then
    rm errors.txt
  fi
  touch errors.txt

  basic_file_dir="basic_idls"
  basic_files=($(find "$basic_file_dir" -name "*.thrift" -type f -print))
  basic_total=${#basic_files[@]}
  echo "starting test"
  for i in "${!basic_files[@]}"; do
    echo "Test [$(($i+1))/$basic_total]:   ${basic_files[$i]}"
    check_cmd ${basic_files[$i]}
  done

  clean_codegen

  if [ ! -s errors.txt ]; then
      # 如果错误输出文件为空，则脚本执行成功
      echo "脚本执行成功，未检测到错误"
  else
      # 如果错误输出文件不为空，则脚本执行失败，打印错误信息
      echo "脚本执行完成，检测到错误!详见 errors.txt"
      cat errors.txt
      exit 1
  fi
}

main