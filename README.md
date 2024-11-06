# gochat

## Preparation

Before starting backend services need to prepare environment.

### ScyllaDB

Create a keyspace

```cassandraql
CREATE KEYSPACE gochat WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};
```

If you know what to do you can change replication factor and strategy.

### MinIO

In WebUI or using CLI create a bucket called `media`

After that change access policy to be able to access files from the web

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": [
                    "*"
                ]
            },
            "Action": [
                "s3:GetBucketLocation"
            ],
            "Resource": [
                "arn:aws:s3:::media"
            ]
        },
        {
            "Effect": "Allow",
            "Principal": {
                "AWS": [
                    "*"
                ]
            },
            "Action": [
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::media/*"
            ]
        }
    ]
}
```