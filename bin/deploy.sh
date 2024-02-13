#!/bin/bash

STAGE=${STAGE:-test}
OUTPUT_BUCKET=sam-artifacts-image-poster-api # make sure this already exist
STACK_NAME=image-poster-api-$STAGE
AWS_PROFILE=${AWS_PROFILE:-image-poster-api}

sam package --profile $AWS_PROFILE --template-file template.yaml --output-template-file packaged.template.yaml --s3-bucket $OUTPUT_BUCKET
sam deploy --profile $AWS_PROFILE --template-file packaged.template.yaml --stack-name $STACK_NAME --capabilities CAPABILITY_AUTO_EXPAND CAPABILITY_IAM CAPABILITY_NAMED_IAM --parameter-overrides Stage=$STAGE --s3-bucket $OUTPUT_BUCKET
