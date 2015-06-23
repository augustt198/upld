package main

import (
    "fmt"
    "time"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
    "encoding/base64"
    "encoding/json"

    "github.com/aws/aws-sdk-go/aws/credentials"
)

func hmac256(key, msg []byte) []byte {
    h := hmac.New(sha256.New, key)
    h.Write(msg)
    return h.Sum(nil)
}

func signatureTime() string {
    t := time.Now()
    return fmt.Sprintf("%04d%02d%02d", t.Year(), t.Month(), t.Day())
}

// policy is a JSON string:
// http://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-HTTPPOSTConstructPolicy.html
func CreateSignature(policy string, creds credentials.Credentials) (string, error) {
    val, err := creds.Get()
    if err != nil {
        return "", err
    }

    policy = base64.StdEncoding.EncodeToString([]byte(policy))

    dateKey := hmac256(
        []byte("AWS4" + val.SecretAccessKey),
        []byte(signatureTime()))
    dateRegionKey := hmac256(dateKey, []byte("us-east-1"))
    dateRegionKeyService := hmac256(dateRegionKey, []byte("s3"))
    signingKey := hmac256(dateRegionKeyService, []byte("aws4_request"))
    signature := hmac256(signingKey, []byte(policy))

    return hex.EncodeToString(signature), nil
}

func CreatePolicy(username string) (string, error) {    
    const layout = "2006-01-02T15:04:05Z" // ISO8601
    expire := time.Now().Add(time.Hour)

    obj := map[string]interface{}{
        "expiration": expire.Format(layout),

        "conditions": []interface{}{
            map[string]interface{}{
                "bucket": config.BucketName,
            },
            []interface{}{
                "starts-width", "$key", username + "/",
            },
            []interface{}{
                "content-length-range", 1, 20000000, // 20MB
            },
            map[string]interface{}{
                "x-amz-algorithm": "AWS4-HMAC-SHA256",
            },
        },
    }

    bytes, err := json.Marshal(obj)
    if err != nil {
        return "", err
    }

    return string(bytes), nil
}
