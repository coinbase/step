# GeoEngineer Resources For Step Function Deployer
# GEO_ENV=development bundle exec geo apply resources/step-deployer.rb

########################################
###           ENVIRONMENT            ###
########################################

env = environment('development') {
  region      ENV.fetch('AWS_REGION')
  account_id  ENV.fetch('AWS_ACCOUNT_ID')
}

########################################
###            PROJECT               ###
########################################
project = project('coinbase', 'step-deployer') {
  environments 'development'
  tags {
    ProjectName "coinbase/step-deployer"
    ConfigName "development"
    DeployWith "step-deployer"
    self[:org] = "coinbase"
    self[:project] = "step-deployer"
  }
}

context = {
  assumed_role_name: "coinbase-step-deployer-assumed",
  assumable_from: [ ENV['AWS_ACCOUNT_ID'] ],
  assumed_policy_file: "#{__dir__}/step_assumed_policy.json.erb"
}

project.from_template('bifrost_deployer', 'step-deployer', {
  lambda_policy_file: "#{__dir__}/step_lambda_policy.json.erb",
  lambda_policy_context: context
})

# The assumed role exists in all environments
project.from_template('step_assumed', 'coinbase-step-deployer-assumed', context)
