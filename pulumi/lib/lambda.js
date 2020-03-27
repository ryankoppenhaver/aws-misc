"use strict";
const aws = require("@pulumi/aws");

const assumeRolePolicy = JSON.stringify({
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
})

exports.lambda = (name, policy, lambdaConfig) => {
  const base = `${name}-lambda`;

  const role = new aws.iam.Role(`${base}-role`, {
    description: `standard role for ${name} lambda`,
    assumeRolePolicy
  })
  
  const rolePolicy = new aws.iam.RolePolicy(`${base}-role_policy`, {
    role,
    policy
  })

  new aws.iam.RolePolicyAttachment(`${base}-rpa_basic_execution`, {
    policyArn: aws.iam.ManagedPolicies.AWSLambdaBasicExecutionRole,
    role
  })

  const lambdaFunction = new aws.lambda.Function(`${base}-func`, {
    role: role.arn,
    ...lambdaConfig
  })

  return {role, rolePolicy, lambdaFunction}
}
