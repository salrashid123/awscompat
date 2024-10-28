
package com.test;

import java.io.FileInputStream;
import java.util.Arrays;
import java.util.List;

import com.amazonaws.regions.Regions;
import com.amazonaws.services.s3.AmazonS3;
import com.amazonaws.services.s3.AmazonS3ClientBuilder;
import com.amazonaws.services.s3.model.ListObjectsV2Result;
import com.amazonaws.services.s3.model.S3ObjectSummary;
import com.google.auth.oauth2.ComputeEngineCredentials;
import com.google.auth.oauth2.GoogleCredentials;
import com.google.auth.oauth2.IdTokenCredentials;
import com.google.auth.oauth2.IdTokenProvider;
import com.google.auth.oauth2.ImpersonatedCredentials;
import com.google.auth.oauth2.ServiceAccountCredentials;
import com.google.awscompat.GCPAWSCredentialProvider;

public class Main {

     private static final String CLOUD_PLATFORM_SCOPE = "https://www.googleapis.com/auth/cloud-platform";
     private static final String credFile = "/home/srashid/gcp_misc/certs/mineral-minutia-820-e9a7c8665867.json";
     private static final String target_audience = "https://sts.amazonaws.com";

     public static void main(String[] args) throws Exception {

          Main tc = new Main();

          // IdTokenCredentials tok = tc.getIDTokenFromComputeEngine(target_audience);

          ServiceAccountCredentials sac = ServiceAccountCredentials.fromStream(new FileInputStream(credFile));
          sac = (ServiceAccountCredentials) sac.createScoped(Arrays.asList(CLOUD_PLATFORM_SCOPE));

          IdTokenCredentials tok = tc.getIDTokenFromServiceAccount(sac, target_audience);

          // String impersonatedServiceAccount =
          // "impersonated-account@project.iam.gserviceaccount.com";
          // IdTokenCredentials tok =
          // tc.getIDTokenFromImpersonatedCredentials((GoogleCredentials)sac,
          // impersonatedServiceAccount, target_audience);

          // AmazonS3 s3 =
          // AmazonS3ClientBuilder.standard().withRegion(Regions.US_EAST_2).build();

          // https://docs.aws.amazon.com/sdk-for-java/v1/developer-guide/credentials.html#credentials-specify-provider

          String roleArn = "arn:aws:iam::291738886548:role/s3webreaderrole";
          AmazonS3 s3 = AmazonS3ClientBuilder.standard().withRegion(Regions.US_EAST_2)
                    .withCredentials(GCPAWSCredentialProvider.builder().roleArn(roleArn).roleSessionName(null)
                              .googleCredentials(tok).build())
                    .build();

          String bucket_name = "mineral-minutia";

          ListObjectsV2Result result = s3.listObjectsV2(bucket_name);
          List<S3ObjectSummary> objects = result.getObjectSummaries();
          for (S3ObjectSummary os : objects) {
               System.out.println("* " + os.getKey());
          }

     }

     public IdTokenCredentials getIDTokenFromServiceAccount(ServiceAccountCredentials saCreds, String targetAudience) {
          IdTokenCredentials tokenCredential = IdTokenCredentials.newBuilder().setIdTokenProvider(saCreds)
                    .setTargetAudience(targetAudience).build();
          return tokenCredential;
     }

     public IdTokenCredentials getIDTokenFromComputeEngine(String targetAudience) {
          ComputeEngineCredentials caCreds = ComputeEngineCredentials.create();
          IdTokenCredentials tokenCredential = IdTokenCredentials.newBuilder().setIdTokenProvider(caCreds)
                    .setTargetAudience(targetAudience)
                    .setOptions(Arrays.asList(IdTokenProvider.Option.FORMAT_FULL, IdTokenProvider.Option.LICENSES_TRUE))
                    .build();
          return tokenCredential;
     }

     public IdTokenCredentials getIDTokenFromImpersonatedCredentials(GoogleCredentials sourceCreds,
               String impersonatedServieAccount, String targetAudience) {
          ImpersonatedCredentials imCreds = ImpersonatedCredentials.create(sourceCreds, impersonatedServieAccount, null,
                    Arrays.asList(CLOUD_PLATFORM_SCOPE), 300);
          IdTokenCredentials tokenCredential = IdTokenCredentials.newBuilder().setIdTokenProvider(imCreds)
                    .setTargetAudience(targetAudience).setOptions(Arrays.asList(IdTokenProvider.Option.INCLUDE_EMAIL))
                    .build();
          return tokenCredential;
     }

}
