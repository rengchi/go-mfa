# Go 语言实现的 TOTP 动态验证码和多因素认证示例

## 下载

```
# https
git clone https://github.com/rengchi/go-mfa.git

# ssh
git clone git@github.com:rengchi/go-mfa.git
```

## 使用

### 准备

```
将文件 `example.config.json` 复制并保存为 `config.json` 文件，修改为正确的数据库配置
```

> 使用 `go run main.go` 需要把 `os.Executable()` 修改为 `os.Getwd()`，建议使用 `go build -o xxx` 并配置环境变量使用

```
PS D:\proj\go\src\github.com\go-mfa> go build -o mfa.exe
PS D:\proj\go\src\github.com\go-mfa> .\mfa.exe
成功连接到数据库！
请输入 TOTP 名称: go_mfa
请输入 TOTP 密钥: VOEBQSF3PXGMGNCLT7NIZNKDUEZME2VAQXHB7ETPIXE3VY2W42NBIHNT2FVUMHJL
当前时间的动态验证码为: 870978
PS D:\proj\go\src\github.com\go-mfa> .\mfa.exe
成功连接到数据库！
请输入 TOTP 名称: go_mfa
当前时间的动态验证码为: 870978
PS D:\proj\go\src\github.com\go-mfa>
```
