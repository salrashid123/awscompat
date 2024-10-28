const AWS = require('aws-sdk');
const { GoogleAuth } = require('google-auth-library');

AWS.config.region = 'us-east-2';

require('./google_compat_credentials.js');

const audience = 'https://sts.amazonaws.com/';
const roleArn = 'arn:aws:iam::291738886548:role/s3webreaderrole';

// https://github.com/googleapis/google-auth-library-nodejs#fetching-id-tokens

async function main() {

  // Init a google IDToken client
  const auth = new GoogleAuth();
  const client = await auth.getIdTokenClient(
    audience
  );

  // 1. Get a raw idtoken and add that to an STS Client
  // const res = await client.getRequestHeaders();
  // const id_token = res.Authorization.replace("Bearer ", "");
  // console.log(id_token);

  // var sts = new AWS.STS();
  // var params = {
  //   RoleArn: roleArn,
  //   RoleSessionName: 'testsession',
  //   WebIdentityToken: id_token,
  // };
  // sts.assumeRoleWithWebIdentity(params, function (err, data) {
  //   if (err) {
  //     console.log(err, err.stack);
  //   } else {
  //     //console.log(data);
  //     AWS.config.update({
  //       accessKeyId: data.Credentials.AccessKeyId,
  //       secretAccessKey: data.Credentials.SecretAccessKey,
  //       sessionToken: data.Credentials.SessionToken,
  //       region: 'us-east-2'
  //     });
  //     var s3 = new AWS.S3();
  //     var params = {
  //       Bucket: 'mineral-minutia',
  //     }
  //     s3.listObjects(params, function (err, data) {
  //       if (err) throw err;
  //       console.log(data);
  //     });
  //   }
  // });

  // 2. Use the IDToken in a WebIdentityCredential() object
  // AWS.config.credentials = new AWS.WebIdentityCredentials({
  //   RoleArn: roleArn,
  //   RoleSessionName: "testsession",
  //   WebIdentityToken: id_token
  // });


  // 3. apply the googleID token Client directly to a wrapped Credential 
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

}

main().catch(console.error);