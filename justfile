# set Windows shell - 设置 Windows shell

set windows-shell := ["powershell.exe", "-c"]

# set `&&` or `;` for different OS - 根据不同的操作系统设置 `&&` 或 `;`

and := if os_family() == "windows" { ";" } else { "&&" }

# load environment from `.env` file - 从 `.env` 文件加载环境变量

set dotenv-load := true

# colors - 颜色

red := '\033[0;31m'
bold := '\033[1m'
normal := '\033[0m'
green := "\\e[32m"
yellow := "\\e[33m"
blue := "\\e[34m"
magenta := "\\e[35m"
grey := "\\e[90m"

#====================================== alias start ============================================#

alias gr := genrpc
alias gm := genmodel
alias t := test
alias dep := dependencies
alias deps := dependencies

#======================================= alias end =============================================#
#===================================== targets start ===========================================#

# default target, excute lint and test - `just` 默认目标，执行代码检查和测试
default: lint test

# build e.g., just build rpc makabaka - 构建
[group('build')]
build target="rpc" proj_name="makabaka":
    @cd {{ if target == "rpc" { target } else { if target == "api" { target } else { "rpc" } } }} {{ and }} cd {{ proj_name }} {{ and }} go build -ldflags "-s -w" .

# generate rpc code. e.g., just genrpc makabaka - 生成 rpc 代码
[confirm("""
Are you sure you want to generate go-zero's rpc and grpc-gateway code?
你确定要生成 go-zero rpc 和 grpc-gateway 代码吗？
input 'Y/N' to continue or exit.
输入 'Y/N' 继续或退出。
""")]
[group('generate')]
genrpc target:
    @goctl rpc protoc --multiple --home {{ join(root, ".goctl") }} {{ join(proto, target) }}.proto -I . -I {{ proto }} --go_out={{ pb }} --go-grpc_out={{ pb }} --zrpc_out={{ join(rpc, target) }} {{ and }} protoc -I . -I {{ proto }} --grpc-gateway_out={{ pb }} {{ join(proto, target) }}.proto

# generate model code. e.g., just genmodel user makabaka - 生成 model 代码
[confirm("""
Are you sure you want to generate go-zero model code?
你确定要生成 go-zero model 代码吗？
input 'Y/N' to continue or exit.
输入 'Y/N' 继续或退出。
""")]
[group('generate')]
genmodel sql_name target="makabaka" *args="":
    @goctl model mysql ddl --home {{ join(root, ".goctl") }} {{ args }} --strict --dir model --src {{ join(root, "sql", target, sql_name) }}.sql

# run e.g., just run rpc makabaka - 运行
[group('dev')]
run target="rpc" proj_name="makabaka":
    @cd {{ if target == "rpc" { target } else { if target == "api" { target } else { "rpc" } } }} {{ and }} cd {{ proj_name }} {{ and }} go run .

# go test
[group('dev')]
test:
    @go test -v {{ join(".", "...") }}

# lint - 代码检查
[group('dev')]
lint: dep-golangci-lint
    @go mod tidy 
    @golangci-lint run

# install dependencies - 安装依赖工具
[group('dependencies')]
dependencies: dep-golangci-lint dep-gofumpt dep-goctl dep-protoc-gen-go dep-protoc-gen-go-grpc dep-protoc-gen-grpc-gateway

# a linter for Go - 一个 Go 语言的代码检查工具
[group('dependencies')]
dep-golangci-lint:
    @go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# a stricter gofmt - 一个更严格的 gofmt
[group('dependencies')]
dep-gofumpt:
    @go install mvdan.cc/gofumpt@latest

[group('dependencies')]
dep-goctl:
    @go install github.com/zeromicro/go-zero/tools/goctl@latest

# this can install protoc, protoc-gen-go, protoc-gen-go-grpc at once - 这个命令可以一次性安装 protoc, protoc-gen-go, protoc-gen-go-grpc
[group('dependencies')]
dep-goctl-env:
    @goctl env check --install --verbose

# open browser and visit https://grpc.io/docs/protoc-installation/ - 打开浏览器访问 https://grpc.io/docs/protoc-installation/
[group('dependencies')]
dep-protoc:
    @echo "{{ yellow }}Please install protoc from https://grpc.io/docs/protoc-installation/"
    @echo "请从 https://grpc.io/docs/protoc-installation/ 安装 protoc{{ normal }}"

[group('dependencies')]
dep-protoc-gen-go:
    @go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

[group('dependencies')]
dep-protoc-gen-go-grpc:
    @go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

[group('dependencies')]
dep-protoc-gen-grpc-gateway:
    @go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

#===================================== targets end ===========================================#
#=================================== variables start =========================================#
# project name - 项目名称

project_name := "makabaka"

# project root directory - 项目根目录

root := justfile_directory()
api := join(root, "api")
rpc := join(root, "rpc")
proto := join(root, "proto")
pb := join(root, "pb")

#=================================== variables end =========================================#
