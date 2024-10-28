import boto3

from botocore.credentials import Credentials

from google.oauth2 import id_token
from google.oauth2 import service_account
import google.auth
import google.auth.transport.requests

# https://github.com/boto/boto3/issues/619#issuecomment-216980368

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

## or ust using default ProcessCredentials
s3_resource = boto3.resource('s3')
bkt = s3_resource.Bucket('mineral-minutia')
for my_bucket_object in bkt.objects.all():
    print(my_bucket_object)
