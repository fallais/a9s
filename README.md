# a9s - Terminal-UI for AWS

**a9s** provides a terminal UI to interact with your AWS resources, developed with **Go** and inspired by [k9s](https://github.com/derailed/k9s).

## Screenshots

![Screenshot 1](assets/screen1.png)

![Screenshot 2](assets/screen3.png)

## Features

- Auto refresh
- Easily select resources
- Switch profile
- Switch region
- S3 : Create, delete and drop (empty) buckets

## Installation

### Linux

```sh
curl -sL https://github.com/fallais/a9s/releases/download/v0.1.0/a9s_0.1.0_linux_arm64.tar.gz
tar -xf a9s_0.1.0_linux_arm64.tar.gz
sudo mv a9s /usr/local/bin/
```

### Windows

Download latest archive from the [Release page](https://github.com/fallais/a9s/releases)

Extract the binary into a specific folder

Add the binary in your PATH

## Resources

- ACM
- EC2
- ECS
- EKS
- Lambda
- RDS
- S3
- ECR
- KMS
- Secrets Manager
- DynamoDB
- Cloudfront
- Cognito