
package com.google.awscompat;

import static com.amazonaws.SDKGlobalConfiguration.AWS_ROLE_ARN_ENV_VAR;
import static com.amazonaws.SDKGlobalConfiguration.AWS_ROLE_SESSION_NAME_ENV_VAR;

import java.io.IOException;

import com.amazonaws.auth.AWSCredentials;
import com.amazonaws.auth.AWSCredentialsProvider;
import com.amazonaws.auth.AWSSessionCredentials;
import com.amazonaws.auth.BasicSessionCredentials;
import com.amazonaws.auth.WebIdentityFederationSessionCredentialsProvider;
import com.google.auth.oauth2.IdTokenCredentials;

public class GCPAWSCredentialProvider implements AWSCredentialsProvider {

    private WebIdentityFederationSessionCredentialsProvider wiCredentialProvider;
    private IdTokenCredentials googleCredentials;
    private final RuntimeException loadException;
    private String roleArn;
    private String roleSessionName;

    public GCPAWSCredentialProvider() {
        this(new BuilderImpl());
    }

    // https://raw.githubusercontent.com/aws/aws-sdk-java/master/aws-java-sdk-core/src/main/java/com/amazonaws/auth/WebIdentityTokenCredentialsProvider.java

    private GCPAWSCredentialProvider(BuilderImpl builder) {
        RuntimeException loadException = null;

        try {

            this.roleArn = builder.roleArn != null ? builder.roleArn : System.getenv(AWS_ROLE_ARN_ENV_VAR);

            this.roleSessionName = builder.roleSessionName != null ? builder.roleSessionName
                    : System.getenv(AWS_ROLE_SESSION_NAME_ENV_VAR);

            if (this.roleSessionName == null) {
                this.roleSessionName = "aws-sdk-java-" + System.currentTimeMillis();
            }

            this.googleCredentials = builder.googleCredentials;

        } catch (RuntimeException e) {

            loadException = e;
        }

        this.loadException = loadException;
    }

    @Override
    public AWSCredentials getCredentials() {
        if (loadException != null) {
            throw loadException;
        }

        try {
            this.googleCredentials.refreshIfExpired();
        } catch (IOException except) {
            throw loadException;
        }

        String idToken = this.googleCredentials.getIdToken().getTokenValue();

        if (this.wiCredentialProvider == null) {
            this.wiCredentialProvider = new WebIdentityFederationSessionCredentialsProvider(idToken, null,
                    this.roleArn);
        }

        AWSSessionCredentials awsCreds = this.wiCredentialProvider.getCredentials();

        BasicSessionCredentials c = new BasicSessionCredentials(awsCreds.getAWSAccessKeyId(),
                awsCreds.getAWSSecretKey(), awsCreds.getSessionToken());
        return c;
    }

    @Override
    public void refresh() {
        System.out.println("refresh()");
    }

    public static GCPAWSCredentialProvider create() {
        return builder().build();
    }

    public static Builder builder() {
        return new BuilderImpl();
    }

    @Override
    public String toString() {
        return getClass().getSimpleName();
    }

    public interface Builder {

        Builder roleArn(String roleArn);

        Builder roleSessionName(String roleSessionName);

        Builder googleCredentials(IdTokenCredentials googleCredentials);

        GCPAWSCredentialProvider build();
    }

    static final class BuilderImpl implements Builder {
        private String roleArn;
        private String roleSessionName;
        private IdTokenCredentials googleCredentials;

        BuilderImpl() {
        }

        @Override
        public Builder roleArn(String roleArn) {
            this.roleArn = roleArn;
            return this;
        }

        public void setRoleArn(String roleArn) {
            roleArn(roleArn);
        }

        @Override
        public Builder roleSessionName(String roleSessionName) {
            this.roleSessionName = roleSessionName;
            return this;
        }

        public void setRoleSessionName(String roleSessionName) {
            roleSessionName(roleSessionName);
        }

        @Override
        public Builder googleCredentials(IdTokenCredentials googleCredentials) {
            this.googleCredentials = googleCredentials;
            return this;
        }

        public void setGoogleCredentials(IdTokenCredentials googleCredentials) {
            googleCredentials(googleCredentials);
        }

        @Override
        public GCPAWSCredentialProvider build() {
            return new GCPAWSCredentialProvider(this);
        }
    }
}