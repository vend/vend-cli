#!/bin/sh

set -e

cd "${WORKING_DIR:-.}"

# Get current tag
CURRENT_TAG=${GITHUB_REF##*tags/}

# If there's no tag, exit neutral
if [ -z "${CURRENT_TAG}" ]; then
    exit 78
fi

# Call the ghr command
ghr \
    -t ${GITHUB_TOKEN} \         # Set Github API Token
    -u ${GITHUB_ACTOR} \         # Set Github username
    -r ${GITHUB_REPOSITORY} \    # Set repository name
    -c ${GITHUB_SHA} \           # Set target commitish, branch or commit SHA
    -n ${CURRENT_TAG} \          # Set release title
    -soft \                      # Stop uploading if the same tag already exists
    ${CURRENT_TAG} ${ARTIFACT}

# Post results back as comment.
COMMENT="#### New Release ${CURRENT_TAG} Published\n\n
${RELEASE_BODY}
"
PAYLOAD=$(echo '{}' | jq --arg body "$COMMENT" '.body = $body')
COMMENTS_URL=$(cat /github/workflow/event.json | jq -r .pull_request.comments_url)
curl -s -S -H "Authorization: token $GITHUB_TOKEN" --header "Content-Type: application/json" --data "$PAYLOAD" "$COMMENTS_URL" > /dev/null

exit 0