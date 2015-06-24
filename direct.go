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

type DirectUpload struct {
    XAmzDate string
    XAmzCredential string
    Policy string
    Signature string
}

func NewDirectUpload(username string, creds credentials.Credentials) (*DirectUpload, error) {
    value, err := creds.Get()
    if err != nil {
        return nil, err
    }
    access := value.AccessKeyID
    t := time.Now().UTC()
    var upload DirectUpload

    signatureTime := fmt.Sprintf("%04d%02d%02d", t.Year(), t.Month(), t.Day())
    upload.XAmzDate = signatureTime + "T000000Z"
    upload.XAmzCredential = fmt.Sprintf(
        "%s/%s/us-east-1/s3/aws4_request",
        access, signatureTime)
    
    policy, err := CreatePolicy(t, username, upload.XAmzCredential, upload.XAmzDate)
    if err != nil {
        return nil, err
    }
    policy = base64.StdEncoding.EncodeToString([]byte(policy))
    upload.Policy = policy
    upload.Signature = CreateSignature(policy, signatureTime, value)

    return &upload, nil
}

// policy is a base64-encoded JSON string:
// http://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-HTTPPOSTConstructPolicy.html
func CreateSignature(policy string, signatureTime string,
    val credentials.Value) string {

    dateKey := hmac256(
        []byte("AWS4" + val.SecretAccessKey),
        []byte(signatureTime))
    dateRegionKey := hmac256(dateKey, []byte("us-east-1"))
    dateRegionKeyService := hmac256(dateRegionKey, []byte("s3"))
    signingKey := hmac256(dateRegionKeyService, []byte("aws4_request"))
    signature := hmac256(signingKey, []byte(policy))

    return hex.EncodeToString(signature)
}

func CreatePolicy(t time.Time, username, xamzcred, xamzdate string) (string, error) {    
    const layout = "2006-01-02T15:04:05Z" // ISO8601
    expire := t.Add(time.Hour)

    obj := map[string]interface{}{
        "expiration": expire.Format(layout),

        "conditions": []interface{}{
            map[string]interface{}{
                "bucket": config.BucketName,
            },
            []interface{}{
                "starts-with", "$key", username + "/",
            },
            []interface{}{
                "starts-with", "$Content-Type", "",
            },
            []interface{}{
                "content-length-range", 1, 20000000, // 20MB
            },
            map[string]interface{}{
                "x-amz-credential": xamzcred,
            },
            map[string]interface{}{
                "x-amz-algorithm": "AWS4-HMAC-SHA256",
            },
            map[string]interface{}{
                "x-amz-date": xamzdate,
            },
        },
    }

    bytes, err := json.Marshal(obj)
    if err != nil {
        return "", err
    }

    return string(bytes), nil
}
