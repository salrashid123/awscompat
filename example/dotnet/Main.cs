
using Amazon.SecurityToken;
using Amazon.SecurityToken.Model;
using Amazon.Runtime;
using Amazon.S3;
using Amazon.S3.Model;
using Amazon;

using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Google.Apis.Auth.OAuth2;
using System.IO;

namespace Program
{
    class Program
    {
        private const string bucketName = "mineral-minutia";
        // Specify your bucket region (an example region is shown).
        private static readonly RegionEndpoint bucketRegion = RegionEndpoint.USEast2;
        private static IAmazonS3 s3Client;
        public static void Main()
        {
            ListObjectsAsync().Wait();
        }

        private static async Task ListObjectsAsync()
        {
            try
            {

                Console.WriteLine("Listing objects stored in a bucket");
                var roleArn = "arn:aws:iam::291738886548:role/s3webreaderrole";
                string CREDENTIAL_FILE_JSON = "/path/to/svc.json";


                // Get the idtoken as string and use it in a standard credential
                // var targetAudience = "https://sts.amazonaws.com/";
                // var idToken = await getGoogleOIDCToken(targetAudience, CREDENTIAL_FILE_JSON);
                // var rawIdToken = await idToken.GetAccessTokenAsync().ConfigureAwait(false);
                // SessionAWSCredentials tempCredentials = await getTemporaryCredentialsAsync(rawIdToken, roleArn, "testsession");
                // Console.WriteLine(tempCredentials.GetCredentials().Token);
                // using (s3Client = new AmazonS3Client(tempCredentials, bucketRegion))

                // or create a usable GoogleCredential to wrap that
                ServiceAccountCredential saCredential;
                using (var fs = new FileStream(CREDENTIAL_FILE_JSON, FileMode.Open, FileAccess.Read))
                {
                    saCredential = ServiceAccountCredential.FromServiceAccountData(fs);
                }                
                var getSessionTokenRequest = new AssumeRoleWithWebIdentityRequest
                {
                    RoleSessionName = "testsession",
                    RoleArn = roleArn
                };
                var cc = new GoogleCompatCredentials(saCredential, "https://sts.amazonaws.com/", getSessionTokenRequest);
                using (s3Client = new AmazonS3Client(cc, bucketRegion))

                //  *****************
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
            }
            catch (AmazonS3Exception s3Exception)
            {
                Console.WriteLine(s3Exception.Message, s3Exception.InnerException);
            }
            catch (AmazonSecurityTokenServiceException stsException)
            {
                Console.WriteLine(stsException.Message, stsException.InnerException);
            }
        }


        public static async Task<OidcToken> getGoogleOIDCToken(string targetAudience, string credentialsFilePath)
        {
            //GoogleCredential gCredential;
            ServiceAccountCredential saCredential;
            //ComputeCredential cCredential;

            using (var fs = new FileStream(credentialsFilePath, FileMode.Open, FileAccess.Read))
            {
                saCredential = ServiceAccountCredential.FromServiceAccountData(fs);
            }

            //cCredential = new ComputeCredential();
            //gCredential = await GoogleCredential.GetApplicationDefaultAsync();
            OidcToken oidcToken = await saCredential.GetOidcTokenAsync(OidcTokenOptions.FromTargetAudience(targetAudience).WithTokenFormat(OidcTokenFormat.Standard)).ConfigureAwait(false);
            return oidcToken;
        }

        private static async Task<SessionAWSCredentials> getTemporaryCredentialsAsync(String idToken, String roleArn, String sessionName)
        {

            using (var stsClient = new AmazonSecurityTokenServiceClient())
            {

                var getSessionTokenRequest = new AssumeRoleWithWebIdentityRequest
                {
                    RoleSessionName = sessionName,
                    RoleArn = roleArn,
                    WebIdentityToken = idToken
                };

                AssumeRoleWithWebIdentityResponse sessionTokenResponse = await stsClient.AssumeRoleWithWebIdentityAsync(getSessionTokenRequest);

                Credentials credentials = sessionTokenResponse.Credentials;

                var sessionCredentials =
                    new SessionAWSCredentials(credentials.AccessKeyId,
                                              credentials.SecretAccessKey,
                                              credentials.SessionToken);
                return sessionCredentials;
            }
        }
    }
}
