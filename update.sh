#!/bin/bash

# 初始化参数变量
TAG=""
COMMIT_MESSAGE=""
INCLUDE_MODULES=""
EXCLUDE_MODULES=""

# 解析参数
while [[ $# -gt 0 ]]; do
    case "$1" in
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -m|--m)
            COMMIT_MESSAGE="$2"
            shift 2
            ;;
        -i|--include)
            INCLUDE_MODULES="$2"
            shift 2
            ;;
        -e|--exclude)
            EXCLUDE_MODULES="$2"
            shift 2
            ;;
        *)
            echo "Invalid options provided: $1"
            exit 1
            ;;
    esac
done

# 检查 commit message 参数是否存在
if [ -z "$COMMIT_MESSAGE" ]; then
    echo "The -m option is required"
    exit 1
fi

# 获取所有子模块目录
ALL_MODULES=()
for dir in */ ; do
    if [ -d "$dir" ]; then
        ALL_MODULES+=("${dir%/}")
    fi
done

# 通用函数用于精确匹配模块名称
filter_modules() {
    local modules=("$@")
    local filter_array=()
    local result=()

    if [ -n "$1" ]; then
        IFS=',' read -r -a filter_array <<< "$1"
        for filter in "${filter_array[@]}"; do
            for module in "${modules[@]}"; do
                if [[ "$module" == "$filter" ]]; then
                    result+=("$module")
                fi
            done
        done
    else
        result=("${modules[@]}")
    fi
    echo "${result[@]}"
}

# 去重函数
remove_duplicates() {
    local arr=("$@")
    local result=()
    local temp
    for i in "${arr[@]}"; do
        if [[ ! " ${result[*]} " =~ " ${i} " ]]; then
            result+=("$i")
        fi
    done
    echo "${result[@]}"
}

# 处理 include 和 exclude 参数
MODULES_TO_PROCESS=($(filter_modules "$INCLUDE_MODULES" "${ALL_MODULES[@]}"))

if [ -n "$EXCLUDE_MODULES" ]; then
    EXCLUDE_ARRAY=($(filter_modules "$EXCLUDE_MODULES" "${ALL_MODULES[@]}"))
    for exclude in "${EXCLUDE_ARRAY[@]}"; do
        MODULES_TO_PROCESS=("${MODULES_TO_PROCESS[@]/$exclude}")
    done
fi

# 去除重复的模块
MODULES_TO_PROCESS=($(remove_duplicates "${MODULES_TO_PROCESS[@]}"))

# 过滤掉空字符串模块
MODULES_TO_PROCESS=("${MODULES_TO_PROCESS[@]}")
MODULES_TO_PROCESS=($(printf "%s\n" "${MODULES_TO_PROCESS[@]}" | awk 'NF'))

# 打印需要执行的子模块列表（JSON格式）
echo "Modules to process: $(printf '%s\n' "${MODULES_TO_PROCESS[@]}" | jq -R . | jq -s .)"

# 获取最新的 tag 并递增版本号
increment_tag() {
    local last_tag=$1
    if [[ $last_tag =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
        local major=${BASH_REMATCH[1]}
        local minor=${BASH_REMATCH[2]}
        local patch=${BASH_REMATCH[3]}
        patch=$((patch + 1))
        echo "v${major}.${minor}.${patch}"
    else
        echo "v1.0.0"
    fi
}

# 初始化一个数组用于存储已处理的子模块和新标签
PROCESSED_MODULES=()

# 检查 tag 是否已存在
if [ -n "$TAG" ]; then
    for MODULE in "${MODULES_TO_PROCESS[@]}"; do
        cd "$MODULE" || { echo "Failed to enter directory $MODULE"; exit 1; }
        if git rev-parse "$TAG" >/dev/null 2>&1; then
            echo "Tag $TAG already exists in module $MODULE. Please provide a different tag."
            exit 1
        fi
        cd ..
    done
fi

# 遍历每个子模块目录
for MODULE in "${MODULES_TO_PROCESS[@]}"; do
    echo "Processing module: $MODULE"

    # 进入子模块目录
    cd "$MODULE" || { echo "Failed to enter directory $MODULE"; exit 1; }
    echo "Current directory: $(pwd)"

#     执行 golangci-lint
#    golangci-lint run
#    if [ $? -ne 0 ]; then
#        echo "Linting failed for module: $MODULE"
#        exit 1
#    fi

    # 检查是否存在 tests 目录，并执行测试
    if [ -d "tests" ]; then
        go test ./tests
        if [ $? -ne 0 ]; then
            echo "Tests failed for module: $MODULE"
            exit 1
        fi
    fi

    # 返回上级目录
    cd ..
done

# 所有子模块都通过了检查，进行提交、打 tag 和 push
for MODULE in "${MODULES_TO_PROCESS[@]}"; do
    echo "Committing and tagging module: $MODULE"

    cd "$MODULE" || { echo "Failed to enter directory $MODULE"; exit 1; }

    git add .
    git commit -m "$COMMIT_MESSAGE"

    # 获取最新的 tag 并递增版本号
    if [ -z "$TAG" ]; then
        last_tag=$(git describe --tags `git rev-list --tags --max-count=1`)
        new_tag=$(increment_tag "$last_tag")
    else
        new_tag="$TAG"
    fi

    git tag "$new_tag"
    git push
    git push origin "$new_tag"

    # 将处理过的模块和新标签存储到数组中
    PROCESSED_MODULES+=("$MODULE => $new_tag")

    cd ..
done

# 打印所有处理过的子模块和对应的新标签
echo "Processed modules and their new tags:"
for entry in "${PROCESSED_MODULES[@]}"; do
    echo "$entry"
done

echo "All modules processed successfully"