#!/bin/sh

for file in ./*
do
    if [ "$file" != "./goal" ] && test -d $file; then
      cd $file
       # 如果有测试用例，执行测试用例
      if [ ! -f "go.work" ]; then
        go work init
        go work use -r ..
      fi

      golangci-lint run

      go mod download

#      git add .
#      git commit -m "sync: 同步更新"

      # 如果有测试用例，执行测试用例
      if test -d "tests"; then
        echo "开始测试$file"
        go test ./tests
      fi
      cd ..
    fi
done

