### Exchange Google and Firebase OIDC tokens for AWS STS


Simple [AWS Credential Provider](https://docs.aws.amazon.com/sdk-for-go/api/aws/credentials/) that uses [Google OIDC tokens](https://github.com/salrashid123/google_id_token).

Essentially, this will allow you to use a google `id_token` for AWS STS `session_token` and then access an aws resource that you've configured an Access Policy for the google identity.  This repo creates an `AWS Credential` derived from a `Google Credential` with the intent of using it for AWS's [IAM Role using External Identities](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_create_for-user_externalid.html).

[Firebase](https://firebase.google.com/) and [Google Cloud Identity Platform](https://cloud.google.com/identity-platform/docs) based `id_tokens` can also be uses for this exchange but is not wrapped into this library (critically since there isn't a golang client library to acquire them).  

>> *NOTE*: the code in this repo is not supported by google.

### Implementations

* [golang](#golang)
* [java](#java)
* [python](#python)
* [dotnet](#dotnet)
* [nodejs](#nodejs)

### References

#### AWS
- [AWS Identity Providers and Federation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers.html)
- [AWS WebIdentityRoleProvider](https://docs.aws.amazon.com/sdk-for-go/api/aws/credentials/stscreds/#WebIdentityRoleProvider)
- [AWS AssumeRoleWithWebIdentity](https://docs.aws.amazon.com/sdk-for-go/api/service/sts/#STS.AssumeRoleWithWebIdentity)
- [aws.credential.Provider](https://godoc.org/github.com/aws/aws-sdk-go/aws/credentials#Provider)

#### Google
- [Authenticating using Google OpenID Connect Tokens](https://github.com/salrashid123/google_id_token)
- [Securely Access AWS Services from Google Kubernetes Engine (GKE)](https://blog.doit-intl.com/securely-access-aws-from-gke-dba1c6dbccba)
- [https://accounts.google.com/.well-known/openid-configuration](https://accounts.google.com/.well-known/openid-configuration)


#### Firebase
- [Firebase Storage and Authorization Rules engine 'helloworld'](https://blog.salrashid.me/posts/firebase_storage_rules/)


### Google OIDC

AWS already supports Google OIDC endpoint out of the box as a provider so the setup is relatively simple: just define an AWS IAM policy that includes google and restrict it with a `Condition` that allows specific external identities as shown below:


- The following definition refers to Role: `arn:aws:iam::291738886548:role/s3webreaderrole`

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "accounts.google.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "accounts.google.com:sub": "100147106996764479085"
        }
      }
    }
  ]
}
```

![images/s3_trust.png](images/s3_trust.png)


To do this by hand, first acquire an ID token (in this case through `gcloud` cli and service account):

```bash
$ gcloud auth activate-service-account --key-file=/path/to/gcp_service_account.json

$ export TOKEN=`gcloud auth print-identity-token --audiences=https://foo.bar`
```

Decode the token using the JWT decoder/debugger at [jwt.io](jwt.io)

The token will show the unique `sub` field that identifies the service account:

```json
{
  "aud": "https://foo.bar",
  "azp": "svc-2-429@mineral-minutia-820.iam.gserviceaccount.com",
  "email": "svc-2-429@mineral-minutia-820.iam.gserviceaccount.com",
  "email_verified": true,
  "exp": 1590898991,
  "iat": 1590895391,
  "iss": "https://accounts.google.com",
  "sub": "100147106996764479085"
}
```

Or using gcloud cli again:

```bash
$ gcloud iam service-accounts describe svc-2-429@mineral-minutia-820.iam.gserviceaccount.com
    displayName: Service Account A
    email: svc-2-429@mineral-minutia-820.iam.gserviceaccount.com
    etag: MDEwMjE5MjA=
    name: projects/mineral-minutia-820/serviceAccounts/svc-2-429@mineral-minutia-820.iam.gserviceaccount.com
    oauth2ClientId: '100147106996764479085'
    projectId: mineral-minutia-820
    uniqueId: '100147106996764479085'
```

Use this `uniqueId` value in the AWS IAM Role policy as shown above.

>> *Note*:  I tried to specify an audience value (`"accounts.google.com:aud": "https://someaud"`) within the AWS policy but that didn't seem to work)
Which means while the `audience` (aud) value is specified in some of the samples here (eg `"https://sts.amazonaws.com/` or `https://foo.bar`) can be anything since its not even currently used in the AWS condition policy)


Export the token and invoke the STS endpoint using the `RoleArn=` value defined earlier

```bash
export TOKEN=eyJhbGciOiJSUzI1...

$ curl -s "https://sts.amazonaws.com/?Action=AssumeRoleWithWebIdentity&DurationSeconds=3600&RoleSessionName=app1&RoleArn=arn:aws:iam::291738886548:role/s3webreaderrole&WebIdentityToken=$TOKEN&Version=2011-06-15&alt=json"
```

You should see AWS `Credential` object in the response
```xml
<AssumeRoleWithWebIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <AssumeRoleWithWebIdentityResult>
    <Audience>svc-2-429@mineral-minutia-820.iam.gserviceaccount.com</Audience>
    <AssumedRoleUser>
      <AssumedRoleId>AROAUH3H6EGKKRVTHVAVB:app1</AssumedRoleId>
      <Arn>arn:aws:sts::291738886548:assumed-role/s3webreaderrole/app1</Arn>
    </AssumedRoleUser>
    <Provider>accounts.google.com</Provider>
    <Credentials>
      <AccessKeyId>ASIAUH3H6EGKPI...</AccessKeyId>
      <SecretAccessKey>EM3Zu4RlDOKGkFPJpceemRqEzfazLk...</SecretAccessKey>
      <SessionToken>FwoGZXIvYXd...</SessionToken>
      <Expiration>2020-05-31T04:23:39Z</Expiration>
    </Credentials>
    <SubjectFromWebIdentityToken>100147106996764479085</SubjectFromWebIdentityToken>
  </AssumeRoleWithWebIdentityResult>
  <ResponseMetadata>
    <RequestId>38dd604d-6ce2-45b3-8e6f-1165ae0e24a1</RequestId>
  </ResponseMetadata>
</AssumeRoleWithWebIdentityResponse>
```

You can manually export the `Credential` in an cli (in this case, to access `s3`)

```bash
export AWS_ACCESS_KEY_ID=ASIAUH3H6EGKIL...
export AWS_SECRET_ACCESS_KEY=+nDF8O2yLDH13ug...
export AWS_SESSION_TOKEN=FwoGZXIvYXd...

$ aws s3 ls mineral-minutia --region us-east-2

    2020-05-29 23:04:07        213 main.py

```

To make this easier, the golang library contained in this repo wraps these steps and provides an AWS `Credential` object for you:


### Usage

There are several ways to exchange GCP credentials for AWS:

You can either delegate the task to get credentials to an external AWS `ProcessCredential` binary or perform the exchange in code as shown in this repo.

#### Process Credentials

In the `ProcessCredential` approach, AWS's client library and CLI will automatically invoke whatever binary is specified in aws's config file.  That binary will acquire a Google IDToken and then exchange it for a `WebIdentityToken` SessionToken from the AWS STS server.  Finally, the binary will emit the tokens to stdout in a specific format that AWS expects.  

For more information, see [Sourcing credentials with an external process](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html) and [AWS Configuration and credential file settings](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)

To use the process credential binary here, first

1. Build the binary

```bash
go build  -o gcp-process-credentials main.go
```

2. Create AWS Config file

create a config file under `~/.aws/config` and specify the path to the binary, the ARN role and optionally the path to a specific gcp service account credential.
```bash
[default]
credential_process = /path/to/gcp-process-credentials  --aws-arn arn:aws:iam::291738886548:role/s3webreaderrole  --gcp-credential-file /path/to/svc.json
```
In the snippet above, i've specified the GCP ServiceAccount Credentials file path.  If you omit that parameter, the binary will use [Google Application Default Credential](https://cloud.google.com/docs/authentication/production) to seek out the appropriate Google Credential Source.   

For example, if you run the binary on GCP VM, it will use the metadata server to get the id_token.   If you specify the ADC Environment varible `GOOGLE_APPLICATION_CREDENTIALS=/path/to.json`, the binary will use the service account specified there

3. Invoke AWS CLI or SDK library

Then either use the AWS CLI or any SDK client in any language.  The library will invoke the process credential for you and acquire the AWS token.

```bash
$ aws s3 ls mineral-minutia --region us-east-2
```

The example output from the binary is just JSON:

```bash
$ gcp-process-credentials  --aws-arn arn:aws:iam::291738886548:role/s3webreaderrole  --gcp-credential-file /path//to/svc.json | jq '.'
{
  "Version": 1,
  "AccessKeyId": "ASIAUH3H6EGKL7...",
  "SecretAccessKey": "YnjWyQFDeeqkRVJQit2uaj+...",
  "SessionToken": "FwoGZ...",
  "Expiration": "2020-06-05T19:24:57+0000"
}
```

### Language Native

The other approach is to exchange the GoogleID token for an AWS one using the AWS Language library itself.  This has the advantage of being "self contained" in that there is no need for an external binary to get installed on the system.

The snippet below demonstrate how to do this in various languages.   For golang and java specifically, the exchange is done a custom AWS credential wrapper which has the distinct advantage of being "managed" just like any other standard AWS Credential type in that any refresh() that is needed on the underlying credential will be self-managed.

At the moment (5/5/20), I havne't been able to figure out how to do this with python..

#### Golang

To use the managed credential in golang, import `"github.com/salrashid123/awscompat/google"` as shown below

```golang
package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sts"
	"io/ioutil"
	"log"
	"time"

	awscompat "github.com/salrashid123/awscompat/google"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

const ()

func main() {

	aud := "https://foo.bar"
	jsonCert := "/path/to/svc.json"

	ctx := context.Background()
	ts, err := idtoken.NewTokenSource(ctx, aud, idtoken.WithCredentialsFile(jsonCert))
	// or on GCE/GKE/Run/GCF, omit the certificate file
	//ts, err := idtoken.NewTokenSource(ctx, aud)
	if err != nil {
		log.Fatalf("unable to create TokenSource: %v", err)
	}

	creds, err := awscompat.NewGCPAWSCredentials(ts, &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:         aws.String("arn:aws:iam::291738886548:role/s3webreaderrole"),
		RoleSessionName: aws.String("app1"),
	})
	if err != nil {
		log.Fatalf("Error creatint Credentials  %v", err)
	}

	sess, err := session.NewSession(&aws.Config{
		Credentials: &creds,
		Region:      aws.String("us-east-2")},
	)
	svcs3 := s3.New(sess)

	sresp, err := svcs3.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String("mineral-minutia")})
	if err != nil {
		log.Fatalf("Error listing objects:  %v", err)
	}

	for _, item := range sresp.Contents {
		fmt.Printf("Name %v  %v\n:", time.Now(), *item.Key)
	}

}
```

Just note 

- `"google.golang.org/api/idtoken"` will only provide `id_tokens` for service accounts.   It does not support user-based id_tokens.


---


#### Java

In Java, the `GCPAWSCredentialProviderl` is included directly in this repo (`java/src/main/java/com/google/awscompat/GCPAWSCredentialProvider.java`) as well as the test app  that acquired googleID tokens from various sources`java/src/main/java/com/test/Main.java`

```java

          // Using default ProcessCredentials
          // AmazonS3 s3 =
          // AmazonS3ClientBuilder.standard().withRegion(Regions.US_EAST_2).build();

          // or
          // Get an IDToken from the source system.
          // IdTokenCredentials tok = tc.getIDTokenFromComputeEngine(target_audience);

          // in this case, its using a service account file:
          ServiceAccountCredentials sac = ServiceAccountCredentials.fromStream(new FileInputStream(credFile));
          sac = (ServiceAccountCredentials) sac.createScoped(Arrays.asList(CLOUD_PLATFORM_SCOPE));

          IdTokenCredentials tok = tc.getIDTokenFromServiceAccount(sac, target_audience);

          // then specify the GCPAWSCredentialProvider.googleCredential() value
          String roleArn = "arn:aws:iam::291738886548:role/s3webreaderrole";
          AmazonS3 s3 = AmazonS3ClientBuilder.standard().withRegion(Regions.US_EAST_2)
                    .withCredentials(GCPAWSCredentialProvider.builder().roleArn(roleArn).roleSessionName(null)
                              .googleCredentials(tok).build())
                    .build();

```

#### Python

For Python, a similar flow to acquire a google ID token manually and then add the raw `SessionToken` value in:

```python
# For other sources of IDTokens, see
#  https://github.com/salrashid123/google_id_token/blob/master/python/googleidtokens.py#L33

def getIdToken():
    svcAccountFile ="/path/to/svc.json"
    target_audience="https://sts.amazonaws.com"
    creds = service_account.IDTokenCredentials.from_service_account_file(
            svcAccountFile,
            target_audience= target_audience)
    request = google.auth.transport.requests.Request()
    creds.refresh(request)
    return creds.token


# Using Specify Credentials and SessionToken
sts_client = boto3.client('sts')

assumed_role_object = sts_client.assume_role_with_web_identity(
    RoleArn="arn:aws:iam::291738886548:role/s3webreaderrole",
    RoleSessionName="AssumeRoleSession1",
    WebIdentityToken=getIdToken(),
    DurationSeconds=900
)
credentials = assumed_role_object['Credentials']

s3_resource = boto3.resource(
    's3',
    aws_access_key_id=credentials['AccessKeyId'],
    aws_secret_access_key=credentials['SecretAccessKey'],
    aws_session_token=credentials['SessionToken'],
)

bkt = s3_resource.Bucket('mineral-minutia')
for my_bucket_object in bkt.objects.all():
    print(my_bucket_object)
```

To Note, the `boto3.resource()` takes the raw value of the `SessionToken` which means once it expires, the `s3_resource` will also fail and not renew.

I havne't been able to figure out how to create a managed `Credential` object in AWS python boto library set that would handle the refresh of the underlying token for you.

#### Dotnet

For dotnet, you can either acquire a GoogleOIDC token and inject it into a static STS client or use the wrapped `GoogleCompatCredentials` object provided in this repo

```csharp
  public class GoogleCompatCredentials : RefreshingAWSCredentials
```

The usage of both modes is shown in `dotnet/Main.cs` while if you just want to use the wrapped version:

You can bootstrap `GoogleCompatCredentials`  using any Google source Credential object that implements its `Google.Apis.Auth.OAuth2.IOidcTokenProvider` interface.  In the case below, its using `ServiceAccountCredential`

```csharp
                ServiceAccountCredential saCredential;
                using (var fs = new FileStream(CREDENTIAL_FILE_JSON, FileMode.Open, FileAccess.Read))
                {
                    saCredential = ServiceAccountCredential.FromServiceAccountData(fs);
                }
                var getSessionTokenRequest = new AssumeRoleWithWebIdentityRequest
                {
                    RoleSessionName = "testsession",
                    RoleArn = roleArn,
                };
                var targetAudience = "https://sts.amazonaws.com/";  // this can be any value (not used in this example)
                var cc = new GoogleCompatCredentials(saCredential, targetAudience, getSessionTokenRequest);
                using (s3Client = new AmazonS3Client(cc, bucketRegion))
                {
                    var listObjectRequest = new ListObjectsRequest
                    {
                        BucketName = bucketName
                    };
                    ListObjectsResponse response = await s3Client.ListObjectsAsync(listObjectRequest);
                    List<S3Object> objects = response.S3Objects;
                    foreach (S3Object o in objects)  {
                       Console.WriteLine("Object  = {0}", o.Key);
                    }
                }
```

#### nodejs

For node, you can either acquire a GoogleOIDC token and inject it into a static STS client or use the wrapped `GoogleCompatCredentials` object provided in this repo


```javascript
const AWS = require('aws-sdk');
const { GoogleAuth } = require('google-auth-library');

AWS.config.region = 'us-east-2';

require('./google_compat_credentials.js');

const audience = 'https://sts.amazonaws.com/';
const roleArn = 'arn:aws:iam::291738886548:role/s3webreaderrole';


  const auth = new GoogleAuth();
  const client = await auth.getIdTokenClient(
    audience
  );


  AWS.config.credentials = new GoogleCompatCredentials({
    RoleArn: roleArn,
    RoleSessionName: "testsession"
  }, client);

  var s3 = new AWS.S3();
  var params = {
    Bucket: 'mineral-minutia',
  }
  s3.listObjects(params, function (err, data) {
    if (err) throw err;
    console.log(data);
  });

```

### Firebase/Identity Platform OIDC

Firebase and [Identity Platform](https://cloud.google.com/identity-platform) can also provide OIDC tokens.  For example, the OIDC `.well-known` endpoint below for a given Firebase Project is discoverable by AWS as an external provider:

For my Firebase/Cloud Identity Platform Project, (called `mineral-minutia-820` below):

- [https://securetoken.google.com/mineral-minutia-820/.well-known/openid-configuration](https://securetoken.google.com/mineral-minutia-820/.well-known/openid-configuration)


and compare that with the full OIDC capabilities of google:

- [https://accounts.google.com/.well-known/openid-configuration](https://accounts.google.com/.well-known/openid-configuration)

But thats enough to get started.  Unfortunately, there doesn't seem to be a golan library to wrap as shown earlier.   The following shows how to get a Firebase id_token using its nodeJS library using plain `firebase.auth().signInWithEmailAndPassword`

```javascript
require("firebase/auth");

const email = "your@email.com";
const password = "yourpassword";
var firebaseConfig = {
    apiKey: "...",
    authDomain: "...",
    projectId: "...",
    storageBucket: "...",
  };
  
firebase.initializeApp(firebaseConfig);
firebase.auth().signInWithEmailAndPassword(email, password).then(result => {
  firebase.auth().currentUser.getIdToken(true).then(function(idToken) {
    console.log(idToken);
  }
}
```
You can find a sample Firebase app that acquires an `id_token` here:

[Firebase Storage and Authorization Rules engine 'helloworld'](https://blog.salrashid.me/posts/firebase_storage_rules/)

Once you have the id_token, decode as done earlier to fin the `sub` value.  THis will be unique to each Firebase project even if the same actual user logs into multiple projects.



First define an external provider:

![images/gcpip_provider.png](images/gcpip_provider.png)

When you specify the endpoint configuration for OIDC, just use the root URL for the discovery endpoint with the firebase projectID: `https://securetoken.google.com/mineral-minutia-820`

Then define a Role with the external provider:  in my case the role was `arn:aws:iam::291738886548:role/cicps3role`

![images/gcpip_role.png](images/gcpip_role.png)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::291738886548:oidc-provider/securetoken.google.com/mineral-minutia-820"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "securetoken.google.com/mineral-minutia-820:aud": "mineral-minutia-820",
          "securetoken.google.com/mineral-minutia-820:sub": "WQoGf9wuwiVtxal5AapvPIfb8Q43"
        }
      }
    }
  ]
}
```

The `sub` value is taken from the decoded `id_token`

```json
{
  "name": "sal a mander",
  "admin": true,
  "groupId": "12345",
  "iss": "https://securetoken.google.com/mineral-minutia-820",
  "aud": "mineral-minutia-820",
  "auth_time": 1590897853,
  "user_id": "WQoGf9wuwiVtxal5AapvPIfb8Q43",
  "sub": "WQoGf9wuwiVtxal5AapvPIfb8Q43",
  "iat": 1590897853,
  "exp": 1590901453,
  "email": "sal@somedomain.com",
  "email_verified": false,
  "firebase": {
    "identities": {
      "email": [
        "sal@somedomain.com"
      ]
    },
    "sign_in_provider": "password"
  }
}
```

As mentioned, the there is no golang library for FireBase so a direct invocation w/ curl will yeld the AWS Credential.

```bash
$  curl -s "https://sts.amazonaws.com/?Action=AssumeRoleWithWebIdentity&DurationSeconds=3600&RoleSessionName=app1&RoleArn=arn:aws:iam::291738886548:role/cicps3role&WebIdentityToken=$TOKEN&Version=2011-06-15&alt=json"
```

```xml
<AssumeRoleWithWebIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <AssumeRoleWithWebIdentityResult>
    <Audience>mineral-minutia-820</Audience>
    <AssumedRoleUser>
      <AssumedRoleId>AROAUH3H6EGKDE7GWMPWA:app1</AssumedRoleId>
      <Arn>arn:aws:sts::291738886548:assumed-role/cicps3role/app1</Arn>
    </AssumedRoleUser>
    <Provider>arn:aws:iam::291738886548:oidc-provider/securetoken.google.com/mineral-minutia-820</Provider>
    <Credentials>
      <AccessKeyId>ASIAUH3H6EGKPJ...</AccessKeyId>
      <SecretAccessKey>Z3S78e6hWYGlub6YlOgz6hYwo81...</SecretAccessKey>
      <SessionToken>FwoGZXIvYXd...</SessionToken>
      <Expiration>2020-05-31T05:11:05Z</Expiration>
    </Credentials>
    <SubjectFromWebIdentityToken>WQoGf9wuwiVtxal5AapvPIfb8Q43</SubjectFromWebIdentityToken>
  </AssumeRoleWithWebIdentityResult>
  <ResponseMetadata>
    <RequestId>5959ed62-13ac-4205-a0f6-811f00bced6b</RequestId>
  </ResponseMetadata>
</AssumeRoleWithWebIdentityResponse>
```

One more note about Firebase/Cloud Identity Platform:  You can use it to define external identities itself (eg, Google, Facebook, AOL, other OIDC, other SAML, etc). 

That means you can chain identities together though Identity Platform.   

Turtles all the way down (or atleast a couple of levels).
