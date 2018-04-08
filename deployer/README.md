# Step Deployer

The Step Deployer is a Step Function that can deploy step functions, so it can recursively deploy itself.

To create the necessary AWS resources you can use GeoEngineer which requires `ruby` and `terraform`:

```bash
bundle install
./geo apply resources/step_deployer.rb
```

```bash
# Use AWS Creds or assume-role
./bootstrap
```

To update the deployer you can use:

```
git pull # pull down new code
./deploy_deployer # call the step-deployer to deploy
```

### Security

Deployers are critical pieces of infrastructure as they may be used to compromise software they deploy. As such, we take security very seriously around the `step-deployer` and answer the following questions:

1. *Authentication*: Who can deploy?
2. *Authorization*: What can be deployed?
3. *Replay* and *Man-in-the-middle (MITM)*: Can some unauthorized person edit or reuse a release to change what is deployed?
4. *Audit*: Who has done what, and when?

#### Authentication

The central authentication mechanisms are the AWS IAM permissions for step functions, lambda, and S3.

By limiting the `lambda:UpdateFunctionCode`, `lambda:UpdateFunctionConfiguration`, `lambda:Invoke*` and `states:UpdateStateMachine` permissions the `step-deployer` function becomes the only way to deploy. Once this is the case, limiting permissions to `states:StartExecution` of the `step-deployer` directly limits who can deploy.

Ensuring the `step-deployer` Lambdas role can only access a single single S3 bucket with:

```
{
  "Effect": "Allow",
  "Action": [
    "s3:GetObject*", "s3:PutObject*",
    "s3:List*", "s3:DeleteObject*"
  ],
  "Resource": [
    "arn:aws:s3:::#{s3_bucket_name}/*",
    "arn:aws:s3:::#{s3_bucket_name}"
  ]
},
{
  "Effect": "Deny",
  "Action": ["s3:*"],
  "NotResource": [
    "arn:aws:s3:::#{s3_bucket_name}/*",
    "arn:aws:s3:::#{s3_bucket_name}"
  ]
},
```

Further restricts who can deploy to those that also can `s3:PutObject` to the bucket.

Who can execute the step function, and who can upload to S3 are the two permissions that guard deploys. Additionally, if you separate those two permissions, you gain extra security, e.g. by only allowing your CI/CD pipe to upload releases, and developers to execute the step function you can ensure only valid builds are ever deployed.

#### Authorization

We use tags and paths to restrict the resources that the `step-deployer` can deploy to.

The lambda function must have a `ProjectName` and `ConfigName` tag that match the release, and a `DeployWith` tag equal to `"step-deployer"`.

Step functions don't support tags, so the path on their role must be must be equal to `/step/<ProjectName>/<ConfigName>/`.

Assets uploaded to S3 are in the path `/<ProjectName>/<ConfigName>` so limiting who can `s3:PutObject` to a path can be used to limit what project-configs they can deploy.

#### Replay and MITM

Each release the client generates a release `release_id`, a `created_at` date, and a SHA256 of the lambda file, and together also uploads the release to S3.

The `step-deployer` will reject any request where the `created_at` date is not recent, the lambdas SHA does not match, or the releases don't match. This means that if a user can invoke the Step function, but not upload to S3 (or vice-versa) it is not possible to deploy old or malicious code.

#### Audit

Working out what happened when is very useful for debugging and security response. Step functions make it easy to see the history of all executions in the AWS console and via API. S3 can log all access to cloud-trail, so collecting from these two sources will show all information about a deploy.

### Continuing Deployment

Some TODOs for the deployer are:

1. Automated rollback on a bad deploy
1. Assume-role sts into other accounts to deploy there, so only one `step-deployer` is needed for many accounts.
