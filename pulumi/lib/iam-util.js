"use strict";

exports.allowPolicy = (config) => {
  const statements = Object.entries(config).map( ([action,resource]) => ({
    "Effect":"Allow",
    "Action": [action],
    //TODO: use flat() when nodejs >= 11
    "Resource": resource instanceof Array ? resource : [resource]
  }))

  const doc = {
    "Version": "2012-10-17",
    "Statement": statements
  }

  return JSON.stringify(doc)
}
