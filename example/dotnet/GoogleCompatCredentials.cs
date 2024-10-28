using System;
using Google.Apis.Auth.OAuth2;

using Amazon.Runtime;

using Amazon;
using Amazon.Runtime.Internal.Util;
using Amazon.SecurityToken;
using Amazon.SecurityToken.Model;
using Amazon.Runtime.Internal;

using System.Globalization;

namespace Program

{
    /// <summary>
    /// AWS Credentials that automatically refresh by calling AssumeRole on
    /// the Amazon Security Token Service.
    /// </summary>
    public class GoogleCompatCredentials : RefreshingAWSCredentials
    {
        private static String defaultTargetAudience = "https://sts.amazonaws.com";
        private RegionEndpoint DefaultSTSClientRegion = RegionEndpoint.USEast1;

        private Logger _logger = Logger.GetLogger(typeof(GoogleCompatCredentials));

        /// <summary>
        /// The credentials of the user that implements Google.Apis.Auth.OAuth2.IOidcTokenProvider.
        /// </summary>
        public Google.Apis.Auth.OAuth2.IOidcTokenProvider SourceCredentials { get; private set; }

        /// <summary>
        /// Sets the audience value for the Google OIDC token.
        /// </summary>
        public String TargetAudience { get; private set; }


        /// <summary>
        /// The Amazon Resource Name (ARN) of the role to assume.  Ignored correctly; its  TODO.!--.
        /// </summary>
        public AssumeRoleWithWebIdentityRequest TargetAssumeRoleRequest { get; private set; }


        /// <summary>
        /// Options to be used in the call to AssumeRole.
        /// </summary>
        public AssumeRoleAWSCredentialsOptions Options { get; private set; }

        /// <summary>
        /// Constructs an GoogleCompatCredentials object.
        /// </summary>
        /// <param name="sourceCredentials">The Google credential that implements Google.Apis.Auth.OAuth2.IOidcTokenProvider.</param>
        /// <param name="targetAudience">The audience value for the GoogleOIDC Token.</param>        
        /// <param name="assumeRoleWithWebIdentityRequest">AssumeRoleWithWebIdentityRequest structure that specifies the Arn, SessionName.</param>
        public GoogleCompatCredentials(Google.Apis.Auth.OAuth2.IOidcTokenProvider sourceCredentials, string targetAudience, AssumeRoleWithWebIdentityRequest wr)
            : this(sourceCredentials, targetAudience, wr, new AssumeRoleAWSCredentialsOptions())
        {
        }

        /// <summary>
        /// Constructs an GoogleCompatCredentials object.
        /// </summary>
        /// <param name="sourceCredentials">The Google credential that implements Google.Apis.Auth.OAuth2.IOidcTokenProvider.</param>
        /// <param name="targetAudience">The audience value for the GoogleOIDC Token.</param>        
        /// <param name="assumeRoleWithWebIdentityRequest">AssumeRoleWithWebIdentityRequest structure that specifies the Arn, SessionName.</param>
        /// <param name="options">Options to be used in the call to AssumeRole. Not implemented!</param>
        public GoogleCompatCredentials(Google.Apis.Auth.OAuth2.IOidcTokenProvider sourceCredentials, string targetAudience, AssumeRoleWithWebIdentityRequest wr, AssumeRoleAWSCredentialsOptions options)
        {
            if (options == null)
            {
                throw new ArgumentNullException("options");
            }

            SourceCredentials = sourceCredentials;
            TargetAudience = defaultTargetAudience;
            if (targetAudience != "") {
                TargetAudience = targetAudience;
            }
            
            TargetAssumeRoleRequest = wr;
            Options = options;
            // Make sure to fetch new credentials well before the current credentials expire to avoid
            // any request being made with expired credentials.
            PreemptExpiryTime = TimeSpan.FromMinutes(5);
        }

        protected override CredentialsRefreshState GenerateNewCredentials()
        {
            var configuredRegion = AWSConfigs.AWSRegion;
            var region = string.IsNullOrEmpty(configuredRegion) ? DefaultSTSClientRegion : RegionEndpoint.GetBySystemName(configuredRegion);

            Amazon.SecurityToken.Model.Credentials cc = null;
            try
            {
                var stsConfig = ServiceClientHelpers.CreateServiceConfig(ServiceClientHelpers.STS_ASSEMBLY_NAME, ServiceClientHelpers.STS_SERVICE_CONFIG_NAME);
                stsConfig.RegionEndpoint = region;

                var stsClient = new AmazonSecurityTokenServiceClient(new AnonymousAWSCredentials());

                OidcToken oidcToken = SourceCredentials.GetOidcTokenAsync(OidcTokenOptions.FromTargetAudience(TargetAudience).WithTokenFormat(OidcTokenFormat.Standard)).Result;

                TargetAssumeRoleRequest.WebIdentityToken = oidcToken.GetAccessTokenAsync().Result;

                AssumeRoleWithWebIdentityResponse sessionTokenResponse = stsClient.AssumeRoleWithWebIdentityAsync(TargetAssumeRoleRequest).Result;

                cc = sessionTokenResponse.Credentials;
                _logger.InfoFormat("New credentials created for assume role that expire at {0}", cc.Expiration.ToString("yyyy-MM-ddTHH:mm:ss.fffffffK", CultureInfo.InvariantCulture));
                return new CredentialsRefreshState(new ImmutableCredentials(cc.AccessKeyId, cc.SecretAccessKey, cc.SessionToken), cc.Expiration);
            }
            catch (Exception e)
            {
                var msg = "Error exchanging Google OIDC token for AWS STS ";
                var exception = new InvalidOperationException(msg, e);
                Logger.GetLogger(typeof(GoogleCompatCredentials)).Error(exception, exception.Message);
                throw exception;
            }

        }
    }
}