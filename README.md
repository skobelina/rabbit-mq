Project contains 2 microservices:
"file-api" - downloads user files
"processing-api" - processes and optimizes images in ".jpg", ".jpeg", ".png" format

1. `$git clone git@github.com:skobelina/rabbit-mq.git`

2. `$cd rabbit-mq`

3. `$make build`

4. `$./file-api/producer` - downloads user files

5. `$./processing-api/consumer` - processes and optimizes images

