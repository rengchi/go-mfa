# Go 语言实现的 TOTP 动态验证码和多因素认证示例

## 下载

```
# https
git clone https://github.com/rengchi/go-mfa.git

# ssh
git clone git@github.com:rengchi/go-mfa.git
```

## 使用

```
PS D:\proj\go\src\github.com\go-mfa> go run .\main.go
成功连接到数据库！
请输入 TOTP 名称: go_mfa
请输入 TOTP 密钥: VOEBQSF3PXGMGNCLT7NIZNKDUEZME2VAQXHB7ETPIXE3VY2W42NBIHNT2FVUMHJL
当前时间的动态验证码为: 870978
PS D:\proj\go\src\github.com\go-mfa> go run .\main.go
成功连接到数据库！
请输入 TOTP 名称: go_mfa
当前时间的动态验证码为: 870978
PS D:\proj\go\src\github.com\go-mfa>
```
