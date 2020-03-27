"use strict";
const pulumi = require("@pulumi/pulumi");
const aws = require("@pulumi/aws");
const awsx = require("@pulumi/awsx");

const {lambda} = require("../lib/lambda")
const {allowPolicy} = require("../lib/iam-util")

// ** TODO ** figure out storage needs (certs, account info), allow write to route53
const policy = TKTK => 

allowPolicy({
  
})

child_process.execSync(`zip -j cert-lambda.zip ${process.env.HOME}/go/src/cert-lambda/cert-lambda`) 
lambda("certs", TKTK, {
  
  code: new pulumi.asset.FileArchive("./cert-lambda.zip"),
  description: "acme client",
  handler: "cert-lambda",
  environment: {
    variables: {
      TKTK
    }
  },
  runtime: "go1.x"
})

const certLambda = new aws.lambda.Function("certLambda", {
  role: ...,
  code: ...,
  description: "periodically update TLS cert from Lets Encrypt",
  handler: "",
})
