# [Upld](http://upld.im)

Upld is an image sharing service written in Go, using
[S3](http://aws.amazon.com/s3/) for storage.

The `thumbgen/` directory contains a thumbnail generator, which
can be triggered by [AWS Lambda](http://aws.amazon.com/lambda/)
in response to uploads.

### Installation

1. Clone the repository
2. Install dependencies (`go get ...`)
3. Compile (`go build`)
