
const AWS = require('aws-sdk');

const { GoogleAuth } = require('google-auth-library');

GoogleCompatCredentials = AWS.util.inherit(AWS.Credentials, {

  constructor: function GoogleCompatCredentials(params, gcpIdTokenClient, clientConfig) {
    AWS.Credentials.call(this);
    this.expired = true;
    this.params = params;
    this.gcpIdTokenClient = gcpIdTokenClient;

    this.params.RoleSessionName = this.params.RoleSessionName || 'web-identity';
    this.data = null;
    this._clientConfig = AWS.util.copy(clientConfig || {});
  },


  refresh: function refresh(callback) {
    this.coalesceRefresh(callback || AWS.util.fn.callback);
  },


  load: function load(callback) {
    var self = this;
    self.createClients();
    self.gcpIdTokenClient.getRequestHeaders().then(res => {
      const id_token = res.Authorization.replace("Bearer ", "");
      this.params.WebIdentityToken = id_token;
      self.service.assumeRoleWithWebIdentity(function (err, data) {
        self.data = null;
        if (!err) {
          self.data = data;
          self.service.credentialsFrom(data, self);
        }
        callback(err);
      });
    });

  },

  createClients: function () {
    if (!this.service) {
      var stsConfig = AWS.util.merge({}, this._clientConfig);
      stsConfig.params = this.params;
      this.service = new AWS.STS(stsConfig);
    }
  }

});

module.exports = GoogleCompatCredentials;