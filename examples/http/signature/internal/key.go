package internal

var (
	// Fingerprint 告诉客户端及服务端使用同一个RSA解密器
	// 客户端告知服务端使用该指纹对应的RSA解密器
	Fingerprint = "12345"

	// PubKey 用于客户端加密
	PubKey = []byte(`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQD7bq4FLG0ctccbEFEsUBuRxkjE
eJ5U+0CAEjJk20V9/u2Fu76i1oKoShCs7GXtAFbDb5A/ImIXkPY62nAaxTGK4KVH
miYbRgh5Fy6336KepLCtCmV/r0PKZeCyJH9uYLs7EuE1z9Hgm5UUjmpHDhJtkAwR
my47YlhspwszKdRP+wIDAQAB
-----END PUBLIC KEY-----`)

	// PriKey 放在服务端用于解密
	PriKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQD7bq4FLG0ctccbEFEsUBuRxkjEeJ5U+0CAEjJk20V9/u2Fu76i
1oKoShCs7GXtAFbDb5A/ImIXkPY62nAaxTGK4KVHmiYbRgh5Fy6336KepLCtCmV/
r0PKZeCyJH9uYLs7EuE1z9Hgm5UUjmpHDhJtkAwRmy47YlhspwszKdRP+wIDAQAB
AoGBANs1qf7UtuSbD1ZnKX5K8V5s07CHwPMygw+lzc3k5ndtNUStZQ2vnAaBXHyH
Nm4lJ4AI2mhQ39jQB/1TyP1uAzvpLhT60fRybEq9zgJ/81Gm9bnaEpFJ9bP2bBrY
J0jbaTMfbzL/PJFl3J3RGMR40C76h5yRYSnOpMoMiKWnJqrhAkEA/zCOkR+34Pk0
Yo3sIP4ranY6AAvwacgNaui4ll5xeYwv3iLOQvPlpxIxFHKXEY0klNNyjjXqgYjP
cOenqtt6UwJBAPw7EYuteVHvHvQVuTbKAaYHcOrp4nFeZF3ndFfl0w2dwGhfzcXO
ROyd5dNQCuCWRo8JBpjG6PFyzezayF4KLrkCQCGditoxHG7FRRJKcbVy5dMzWbaR
3AyDLslLeK1OKZKCVffkC9mj+TeF3PM9mQrV1eDI7ckv7wE7PWA5E8wc90MCQEOV
MCZU3OTvRUPxbicYCUkLRV4sPNhTimD+21WR5vMHCb7trJ0Ln7wmsqXkFIYIve8l
Y/cblN7c/AAyvu0znUECQA318nPldsxR6+H8HTS3uEbkL4UJdjQJHsvTwKxAw5qc
moKExvRlN0zmGGuArKcqS38KG7PXZMrUv3FXPdp6BDQ=
-----END RSA PRIVATE KEY-----`)

	// Key 用于客户端加密(EcbEncrypt)、服务端解密(EcbDecrypt)
	Key = []byte("q4t7w!z%C*F-JaNdRgUjXn2r5u8x/A?D")
)