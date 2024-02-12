# image-poster-api

## Overview

A prototype for an API that allows clients to post an images with captions and clients can create a comment on each post.

### Constraints

- 1 post can only have 1 image
- A post can have multiple comments
- Minimum throughput: 100 RPS
- Cater for client's unstable internet connection

### Tech stack

- AWS Lambda
- AWS S3
- AWS DynamoDB
- AWS API Gateway

## Flows

### Creating Post

![create post flow diagram](docs/create-post-flow.png)

### Creating Comment

![create comment flow diagram](docs/create-comment-flow.png)

### Deleting Comment

![delete comment flow diagram](docs/delete-comment-flow.png)

### Retrieving Posts

![get post flow diagram](docs/get-flow.png)

## Design Limitations

The design prioritizes low latency, high throughput, and could cater for users with unstable internet connections. However, to achieve these constraints, it introduces some tradeoffs which are listed below:

- The text payload (e.g. caption, user) is only limited to 2 KB (around ~2k characters). The limitation is on S3 metadata.
- Increasing the number of comments that comes with the GET /posts response will not be trivial because the data is being duplicated in the Post entity to reduce latency. _e.g. bumping from latest 2 comments to latest 5 comments_
