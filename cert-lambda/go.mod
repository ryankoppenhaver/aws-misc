module cert-lambda

go 1.13

require (
	github.com/aws/aws-lambda-go v1.13.3
	github.com/go-acme/lego v0.0.0-00010101000000-000000000000
	github.com/go-acme/lego/v3 v3.3.0 // indirect
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550
	gopkg.in/square/go-jose.v2 v2.3.1
)

replace github.com/go-acme/lego => /home/rlk/go/src/github.com/go-acme/lego
