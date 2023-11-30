#!/bin/bash

# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
APP_NAME="ccx-data-pipeline"  # name of app-sre "application" folder this component lives in
# NOTE: ccx-notification-writer contains deployment for multiple services
#       for pull requests we need latest git PR version of these components to be
#       deployed to ephemeral env and overriding resource template --set-template-ref.
#       Using multiple components name in COMPONENT_NAME forces bonfire to use the
#       git version of clowdapp.yaml(or any other) file from the pull request.
COMPONENT_NAME="ccx-notification-writer ccx-notification-db-cleaner"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
IMAGE="quay.io/cloudservices/ccx-notification-writer"
COMPONENTS="ccx-notification-writer ccx-notification-service ccx-notification-db-cleaner"  # space-separated list of components to laod
COMPONENTS_W_RESOURCES="ccx-notification-writer"  # component to keep
CACHE_FROM_LATEST_IMAGE="true"

export IQE_PLUGINS="ccx"
export IQE_MARKER_EXPRESSION=""
# Workaround: There are no cleaner specific integration tests. Check that the service loads and iqe plugin works.
export IQE_FILTER_EXPRESSION="test_plugin_accessible"
export IQE_REQUIREMENTS_PRIORITY=""
export IQE_TEST_IMPORTANCE=""
export IQE_CJI_TIMEOUT="30m"


function build_image() {
    source $CICD_ROOT/build.sh
}

function deploy_ephemeral() {
    source $CICD_ROOT/deploy_ephemeral_env.sh
}

function run_smoke_tests() {
   # Workaround: cji_smoke_test.sh requires only one component name. Fallback to only one component name.
    export COMPONENT_NAME="ccx-notification-writer"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
    source $CICD_ROOT/cji_smoke_test.sh
    source $CICD_ROOT/post_test_results.sh  # publish results in Ibutsu
}


# Install bonfire repo/initialize
CICD_URL=https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd
curl -s $CICD_URL/bootstrap.sh > .cicd_bootstrap.sh && source .cicd_bootstrap.sh
echo "creating PR image"
build_image

echo "deploying to ephemeral"
deploy_ephemeral

echo "running PR smoke tests"
run_smoke_tests
