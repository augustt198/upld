# [Upld](http://upld.im)

Upld is an image sharing service written in Go, using
[S3](http://aws.amazon.com/s3/) for storage.

The `thumbserv/` directory contains a thumbnail generator, which
receives messages from [SQS](http://aws.amazon.com/sqs/).

### Installation

1. Clone the repository
2. Install dependencies (`go get ...`)
3. `go build` in the root and `thumbserv/` directories
