package service

// trustedPublicKey 产品硬编码的可信公钥（编译时嵌入，不可被运行时篡改）
// 此公钥必须与 License Server 中对应产品的公钥一致
// 更换产品密钥对后需同步更新此处并重新编译
const trustedPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzt3WJhLqYKKSyHj62rhn
bvzUvMLsPMP6lCXUiugkFskeAjrKhjqgzeNuiZKr4Wa+x+sizdi+E0EnH+pUplHj
XPj42fQR20pUiRdTZD2YzEnaINI/fhg4lhNrheH4QUpjRHUm5b9b/avLV4gNQaLp
niSIpLpjkG1/nvu3IR7hZJz9U7a3F1x3Y6KfFyUuZcatNmhlWEG/Mcnyh6VWxaoy
Obywlh323fAH3J80xApXkAM9xisfuiB2yp59vYrHjLgmjnEE30yN1+9nRQRusx78
PJ0kl13x6an6WpNS1jA7plDG27FraNJ/Nr9CmPH5NYTdPcOwEqv32MSGZV3BO420
UQIDAQAB
-----END PUBLIC KEY-----`
