# ec2-route53

EC2 가변 공인 IP를 Route 53 도메인에 연결시켜주는 Go로 작성한 Lambda 프로그램이다.

EventBridge 이용해 EC2 상태 변화 이벤트에 대한 트리거를 걸고, 그에 따른 Lambda가 실행되는 식이다.


Windows Powershell 기준, 이 코드 빌드 방법은 다음과 같다.

### `build-lambda-zip` 다운로드

```
go install github.com/aws/aws-lambda-go/cmd/build-lambda-zip@latest
```

### Go 빌드하기

```
$env:GOOS = "linux"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "0"
go build -tags lambda.norpc -o bootstrap main.go
build-lambda-zip -o myFunction.zip bootstrap
```

### Lambda 생성 혹은 갱신

최초 생성

```
aws lambda create-function --function-name myFunction3 --runtime provided.al2023 --handler bootstrap --architectures arm64 --role [Role ARN] --zip-file fileb://myFunction.zip
```

Role에 부여해야할 권한은 다음과 같다. (대충 정함)

1. `AmazonEC2ReadOnlyAccess` - EC2 정보(IP 주소, 태그 등) 읽어오기 위해
2. `AmazonRoute53FullAccess` - 도메인 설정 조작하기 위해
3. `AWSLambdaBasicExecutionRole` - Lambda 실행되기 위해

갱신

```
aws lambda update-function-code --function-name myFunction3 --zip-file fileb://myFunction.zip
```

### 지정해야 할 EC2 태그

1. `DomainName` - 예: blog.example.com
2. `HostedZoneId` - Route 53 가면 볼 수 있다.

### EventBridge 설정

https://repost.aws/knowledge-center/ec2-email-instance-state-change 참고하도록...
