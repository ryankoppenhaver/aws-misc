"use strict";
const pulumi = require("@pulumi/pulumi");
const aws = require("@pulumi/aws");
const awsx = require("@pulumi/awsx");

const config = new pulumi.Config()

const stackName = pulumi.getStack()
const stackEnv = (()=>{
  try { 
    return stackName.match(/^\w+\.(\w+)$/)[1]
  } catch (e) {
    throw `Stack name "${stackName}" does not match expected pattern`
  }
})()

const infra = new pulumi.StackReference("infra.prod") //only prod for infra
const zone = infra.getOutputSync("zone")

//TODO extract these somewhere?
const sshKeyName = "rlk@descartes-20180930"

const bucket = new aws.s3.Bucket("www-data", {
  acl: "private"
});
exports.bucketName = bucket.id;

const role = new aws.iam.Role("www", {
  description: "Has read access to some S3 Buckets",
  assumeRolePolicy: JSON.stringify({
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "Service": "ec2.amazonaws.com"
        },
      "Action": "sts:AssumeRole"
      }
    ]
  })
})

const addS3Read = (role, bucket) => {

  let bucketName, bucketId
  if (typeof bucket === "string") {
    bucketName = bucket
    bucketId = pulumi.output(bucket)
  } else {
    bucketName = bucket.__name
    bucketId = bucket.id
  }

  //TODO need a clear convention for constructing names
  const rpName = `${role.__name}-read-${bucketName}`

  return new aws.iam.RolePolicy(rpName, {
    role: role,
    policy: bucketId.apply( id => JSON.stringify({
      "Version": "2012-10-17",
      "Statement": [
          {
              "Effect": "Allow",
              "Action": ["s3:ListBucket"],
              "Resource": [`arn:aws:s3:::${id}`]
          },
          {
              "Effect": "Allow",
              "Action": ["s3:GetObject"],
              "Resource": [`arn:aws:s3:::${id}/*`]
          }
      ]
    }))
  })
}

addS3Read(role, "rlk-general-private") //for letsencrypt creds
addS3Read(role, bucket)

new aws.iam.RolePolicyAttachment("www-getcert", {
  policyArn: infra.getOutputSync("instanceCertPolicyARN"),
  role
})

const instanceProfile = new aws.iam.InstanceProfile("www", {
    role: role.name
})

//TODO pull from amazon?  force consistency w/ git sha?
const amiId = require("../../manifest.json").builds
  .sort( (a,b) => b.build_time - a.build_time )[0]
  .artifact_id
  .replace(/^[\w-]+:/, '') //strip leading "us-west-2:"


const instance = new aws.ec2.Instance("www", {
  ami: amiId,
  instanceType: "t2.micro",
  iamInstanceProfile: instanceProfile.name,
  keyName: sshKeyName,
  userData: pulumi.all([bucket.id, role.name]).apply(([b, r]) => JSON.stringify({
    wwwDataBucket: b,
    wwwDataRole: r,
    
    caLambdaARN: infra.getOutputSync("caLambdaARN"),
  }))
})

const eip = new aws.ec2.Eip("www", {instance})
exports.publicIp = eip.publicIp


const aRecordFor = (name, ip) => {
  new aws.route53.Record(name, {
    name: name,
    type: "A",
    ttl: 1800,
    records: [ip],
    zoneId: zone.id
  })
}
 
if (stackEnv === "prod") {
  aRecordFor(zone.name, eip.publicIp)
  aRecordFor(`www.${zone.name}`, eip.publicIp)
} else {
  aRecordFor(`${stackEnv}.${zone.name}`, eip.publicIp)
}
