module github.com/future-bots/producer

go 1.22.2

require (
	github.com/future-bots/platform v0.0.0
	github.com/segmentio/kafka-go v0.4.43
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
)

replace github.com/future-bots/platform => ../../libs/go/platform

replace github.com/future-bots/proto => ../../proto/gen/go
