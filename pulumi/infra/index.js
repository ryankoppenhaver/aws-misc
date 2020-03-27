"use strict";
const child_process = require('child_process')

const pulumi = require("@pulumi/pulumi");
const aws = require("@pulumi/aws");
const awsx = require("@pulumi/awsx");

const {lambda} = require("../lib/lambda")
const {allowPolicy} = require("../lib/iam-util")

const domain = "bitflip.space"
const zone = new aws.route53.Zone("bitflip.space", {
  name: domain
})
exports.zone = zone

// CA key is manually uploaded



const caBucket = new aws.s3.Bucket("certificate-authority", {
  acl: 'private'
})
const CA_ROOT_PATH = "ca/v1"
const caFullRoot = pulumi.interpolate `${caBucket.id}/${CA_ROOT_PATH}`
const caKeyFullPath = pulumi.interpolate `${caFullRoot}/ca-key`
const caOutputFullPrefix = pulumi.interpolate `${caFullRoot}/signed-certs`

const caPolicy = pulumi.all([caKeyFullPath, caOutputFullPrefix]).apply( ([kp,op]) => allowPolicy({
  "iam:GetRole": "*",
  "s3:GetObject":`arn:aws:s3:::${kp}`, //read ca key
  "s3:PutObject":`arn:aws:s3:::${op}/*`//write signed certs
}))

child_process.execSync(`zip -j ca-lambda.zip ${process.env.HOME}/go/src/ssh-signer/ssh-signer`) 

exports.caLambdaARN = lambda("ca", caPolicy, {
  code: new pulumi.asset.FileArchive("./ca-lambda.zip"),
  description: "SSH CA using S3 w/ IAM policy to securely deliver certs",
  handler: "ssh-signer",
  environment: {
    variables: {
      CA_KEY_URI: pulumi.interpolate `s3://${caKeyFullPath}`,
      OUTPUT_URI_PREFIX: pulumi.interpolate `s3://${caOutputFullPrefix}`
    }
  },
  runtime: "go1.x"
}).lambdaFunction.arn

const instanceCertPolicy = new aws.iam.Policy("instanceCertPolicy", {
  description: "allows instances to access their own CA-signed certs",
  policy: pulumi.all([caOutputFullPrefix, exports.caLambdaARN]).apply( ([op,arn]) => allowPolicy({
    "lambda:InvokeFunction": arn,
    "s3:GetObject": `arn:aws:s3:::${op}/\${aws:userid}/certificate`
  }))
})
exports.instanceCertPolicyARN = instanceCertPolicy.arn
