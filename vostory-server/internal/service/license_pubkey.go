package service

// trustedPublicKey 产品硬编码的可信公钥（编译时嵌入，不可被运行时篡改）
// 此公钥必须与 License Server 中对应产品的公钥一致
// 更换产品密钥对后需同步更新此处并重新编译
const trustedPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwpgyey4fDK0ZcfiH1nCv
t9Jhcj7RdVS00esSrgaFMUV/fqUCo1zAWBQbSSXiYv92bgTxFiaZ8rB2fSnAU02c
3+ekSqQv/rdsHecDQVVkMdEFg/phDq8oN95VYO737ll4tRTEBaSzYQyNMXKlbsGT
KpjoN+2vWh7MW+9KenPqQ32krQKz1N3cH7GU47LFXDs0kmEgVRo9f2hmfQLd3qY7
WDdoUru4qDyHEQpUYL6tN7VXkVz7OhXLujQ5R0vsyy1LmulXlO6msynmqtlk5Fmo
i3EnLvgCyvBptRIbtRuUlP165F3qm25qCAqoe9XTvCzLmCYQA7WquejtFGG6aStL
ZwIDAQAB
-----END PUBLIC KEY-----`
